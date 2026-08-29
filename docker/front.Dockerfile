FROM php:8.4-apache

RUN apt-get update && apt-get install -y --no-install-recommends \
        libcurl4-openssl-dev libonig-dev libzip-dev openssl \
    && docker-php-ext-install curl mbstring zip \
    && a2enmod ssl rewrite \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /etc/apache2/ssl \
    && openssl req -x509 -nodes -days 825 -newkey rsa:2048 \
        -keyout /etc/apache2/ssl/nmw.key -out /etc/apache2/ssl/nmw.crt \
        -subj "/C=FR/O=NoMoreWaste/CN=nomorewaste.local" \
        -addext "basicConstraints=critical,CA:FALSE" \
        -addext "keyUsage=digitalSignature,keyEncipherment" \
        -addext "subjectAltName=DNS:nomorewaste.local,DNS:localhost,IP:127.0.0.1,IP:192.168.100.128"

COPY docker/front-vhost.conf /etc/apache2/sites-available/000-default.conf
COPY docker/front-ssl.conf /etc/apache2/sites-available/nmw-ssl.conf
RUN a2ensite nmw-ssl \
    && echo "ServerName nomorewaste.local" >> /etc/apache2/apache2.conf

WORKDIR /var/www/html
COPY index.html ./
COPY generic.css ./
COPY php/erreur.php ./
COPY php/frontoffice ./frontoffice
COPY php/backoffice ./backoffice
COPY php/inc ./inc
COPY php/lang ./lang

EXPOSE 80 443
