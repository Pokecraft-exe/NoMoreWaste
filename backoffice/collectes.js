    const collectes = [
      { id: 0, lieu: "Paris", date: "2026-09-15", partenaire: "Croix-Rouge" },
      { id: 1, lieu: "Lyon", date: "2026-09-15", partenaire: "Secours Populaire" },
      { id: 2, lieu: "Marseille", date: "2026-09-20", partenaire: "Restos du Coeur" },
      { id: 3, lieu: "Toulouse", date: "2026-09-22", partenaire: "Croix-Rouge" },
      { id: 4, lieu: "Nantes", date: "2026-10-01", partenaire: "Banque Alimentaire" }
    ];

    const lieuInput = document.getElementById("lieu");
    const dateInput = document.getElementById("date");
    const partenaireSelect = document.getElementById("partenaire");
    const results = document.getElementById("results");
    const searchBtn = document.getElementById("searchBtn");
    const resetBtn = document.getElementById("resetBtn");

    function initPartenaires() {
      const uniques = [...new Set(collectes.map(c => c.partenaire))];
      uniques.forEach(name => {
        const option = document.createElement("option");
        option.value = name;
        option.textContent = name;
        partenaireSelect.appendChild(option);
      });
    }

    function render(list) {
      results.innerHTML = "";

      if (list.length === 0) {
        const empty = document.createElement("div");
        empty.className = "empty";
        empty.textContent = "Aucune collecte trouvée.";
        results.appendChild(empty);
        return;
      }

      list.forEach(item => {
        const card = document.createElement("div");
        card.className = "card";
        card.innerHTML = `<a href="/collecte?id=${item.id}">
          <strong>${item.lieu}</strong><br>
          Date: ${item.date}<br>
          Partenaire: ${item.partenaire}
        </a>
        `;
        results.appendChild(card);
      });
    }

    function rechercher() {
      const lieu = lieuInput.value.trim().toLowerCase();
      const date = dateInput.value;
      const partenaire = partenaireSelect.value;

      const filtered = collectes.filter(item => {
        const matchLieu = !lieu || item.lieu.toLowerCase().includes(lieu);
        const matchDate = !date || item.date === date;
        const matchPartenaire = !partenaire || item.partenaire === partenaire;
        return matchLieu && matchDate && matchPartenaire;
      });

      render(filtered);
    }

    function resetFilters() {
      lieuInput.value = "";
      dateInput.value = "";
      partenaireSelect.value = "";
      render(collectes);
    }

    searchBtn.addEventListener("click", rechercher);
    resetBtn.addEventListener("click", resetFilters);

    initPartenaires();
    render(collectes);