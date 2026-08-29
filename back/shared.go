package main

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// All overridable with environment variables at deploy time. If
// DATABASE_AUTH_URL is not set, the others fall back to it, so a
// single-database setup only needs one env var.
var (
	DATABASE_AUTH_URL   = getEnv("DATABASE_AUTH_URL", "postgres://postgres:035464@localhost:5432/database")
	DATABASE_PUBLIC_URL = getEnv("DATABASE_PUBLIC_URL", DATABASE_AUTH_URL)
	DATABASE_ADMIN_URL  = getEnv("DATABASE_ADMIN_URL", DATABASE_AUTH_URL)
)

var DATA_DIR = getEnv("DATA_DIR", "./DATA")

var (
	telephoneRegexp  = regexp.MustCompile(`^\+[0-9]{11}$`)
	codePostalRegexp = regexp.MustCompile(`^[0-9]{5}$`)
)

// Types d'utilisateur en dur (miroir de compte.type_utilisateur en base et
// de la constante equivalente cote oauth). Il n'y a plus de table de roles :
// toutes les autorisations de l'API se decident en comparant directement
// token.UserType a l'une de ces valeurs.
const (
	UserTypeVisiteur       = "visiteur"
	UserTypeBenevole       = "benevole"
	UserTypeAdherent       = "adherent"
	UserTypeCommercant     = "commercant"
	UserTypeResponsable    = "responsable"
	UserTypeAdministrateur = "administrateur"
)

var validUserTypes = []string{
	UserTypeVisiteur,
	UserTypeBenevole,
	UserTypeAdherent,
	UserTypeCommercant,
	UserTypeResponsable,
	UserTypeAdministrateur,
}

func isValidUserType(userType string) bool {
	return contains(validUserTypes, userType)
}

// selfServiceUserTypes is what /oauth/v3/inscription accepts directly: the
// use case diagram only exposes "S'inscrire comme Commercant / Adherant /
// Visiteur" on the front-office. Becoming benevole goes through a
// candidature reviewed by staff, and responsable/administrateur accounts are
// only ever created by an administrateur - never by public registration.
var selfServiceUserTypes = []string{
	UserTypeVisiteur,
	UserTypeAdherent,
	UserTypeCommercant,
}

func isSelfServiceUserType(userType string) bool {
	return contains(selfServiceUserTypes, userType)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// IntrospectionPayload mirrors the JSON returned by the oauth service's
// /oauth/v3/introspect. UserType is the account's type_utilisateur - the
// oauth service puts it directly in "scope" as a single value, never a list
// of roles, and every authorization check in this API compares against it.
type IntrospectionPayload struct {
	Active   bool   `json:"active"`
	UserType string `json:"scope"`
	CompteID int    `json:"client_id"`
	Username string `json:"username"`
}

func isAdministrateur(token *IntrospectionPayload) bool {
	return token.UserType == UserTypeAdministrateur
}

// isStaff covers both permanent staff (responsable) and administrateur -
// the two types allowed to run the association's back-office.
func isStaff(token *IntrospectionPayload) bool {
	return token.UserType == UserTypeAdministrateur || token.UserType == UserTypeResponsable
}

func isBenevole(token *IntrospectionPayload) bool {
	return token.UserType == UserTypeBenevole
}

func isAdherent(token *IntrospectionPayload) bool {
	return token.UserType == UserTypeAdherent
}

func isCommercant(token *IntrospectionPayload) bool {
	return token.UserType == UserTypeCommercant
}

// tryAuth validates the caller's bearer token and returns the resulting
// IntrospectionPayload. Oauth and API now live in the same process, so this
// validates the token directly against the database (see validateToken)
// instead of round-tripping through /oauth/v3/introspect over HTTP - that
// endpoint still exists (oauth.go) for external/spec-compliant callers, but
// nothing in this service needs it internally any more. tryAuth writes the
// error response itself and returns a payload with Active=false when
// anything goes wrong, so callers only need to check token.Active.
func tryAuth(w http.ResponseWriter, r *http.Request) *IntrospectionPayload {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		writeAPIError(w, http.StatusBadRequest, "invalid_client", "The client failed to authenticate with the authorization server.")
		return &IntrospectionPayload{}
	}

	payload := validateToken(r.Context(), strings.TrimPrefix(authHeader, "Bearer "))
	if !payload.Active || payload.CompteID == 0 {
		writeAPIError(w, http.StatusUnauthorized, "forbidden", "verify your access token")
		return &IntrospectionPayload{}
	}

	return payload
}

// tryOptionalAuth is like tryAuth but never writes an error response - used
// where a bearer token is welcome but not required (e.g. a guest submitting
// a support ticket).
func tryOptionalAuth(r *http.Request) *IntrospectionPayload {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return &IntrospectionPayload{}
	}
	return validateToken(r.Context(), strings.TrimPrefix(authHeader, "Bearer "))
}

func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeAPIError(w http.ResponseWriter, status int, errCode string, description string) {
	writeJSON(w, status, map[string]any{
		"error":             errCode,
		"error_description": description,
	})
}

// parsePagination reads the from/size query params required by every list
// endpoint in this API.
func parsePagination(r *http.Request) (from int, size int, err error) {
	fromRaw := r.URL.Query().Get("from")
	sizeRaw := r.URL.Query().Get("size")
	if fromRaw == "" || sizeRaw == "" {
		return 0, 0, errMissingPagination
	}

	from, err = strconv.Atoi(fromRaw)
	if err != nil || from < 0 {
		return 0, 0, errInvalidFrom
	}

	size, err = strconv.Atoi(sizeRaw)
	if err != nil || size <= 0 {
		return 0, 0, errInvalidSize
	}

	return from, size, nil
}

var (
	errMissingPagination = &paginationError{"from and size query params are required"}
	errInvalidFrom       = &paginationError{"from must be a non-negative integer"}
	errInvalidSize       = &paginationError{"size must be a strictly positive integer"}
)

type paginationError struct{ message string }

func (e *paginationError) Error() string { return e.message }
