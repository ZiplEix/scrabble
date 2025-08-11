# Scrabble — Monorepo (Backend API + Frontend)

Monorepo du jeu de **Scrabble** en ligne :

* **API** (Go + Echo + PostgreSQL + Goose migrations)
* **Frontend** (SvelteKit + TypeScript + Tailwind + PWA/Push)

> Chaque sous‑projet possède son README dédié :
>
> * [API — README](./api)
> * [Frontend — README](./frontend)

---

## Aperçu

* Parties multijoueurs : plateau 15×15, règles FR, scores, cases spéciales.
* Authentification JWT, rôles (admin).
* Historique des coups, suggestions d’utilisateurs, système de « reports » (tickets).
* Notifications Web Push (VAPID).
* Migrations SQL (Goose).

---

## Arborescence

```
.
├── Makefile                # commandes utiles (db, migrate, dev)
├── docker-compose.yml      # Postgres (et services optionnels)
├── api/                    # backend Go
├── frontend/               # app SvelteKit
└── migrations/             # scripts Goose (SQL + runners up/down)
```

---

## Prérequis

* **Docker** & **docker-compose** (recommandés pour la base de données)
* **Go** (version selon `api/go.mod`)
* **Node.js 20+** et **npm** (frontend)

---

## Démarrage rapide (5 minutes)

1. **Cloner** le repo et se placer à la racine.

2. **Configurer les variables d’environnement** :

   * API :

     ```bash
     cp api/.env.example api/.env
     # Éditez POSTGRES_URL, JWT_SECRET, VAPID_PUBLIC_KEY/PRIVATE_KEY
     ```
   * Frontend :

     ```bash
     cp frontend/.env.example frontend/.env  # si présent, sinon créez‑le
     # Variables minimales :
     # PUBLIC_API_BASE_URL="http://localhost:8888/"
     # PUBLIC_VAPID_PUBLIC_KEY="<clé VAPID publique>"
     ```

3. **Lancer la base de données** :

   ```bash
   make db               # démarre Postgres via docker‑compose
   # ou : docker compose up -d db
   ```

4. **Appliquer les migrations** :

   ```bash
   make migrate-up
   ```

5. **Démarrer l’API en dev** :

   ```bash
   make air              # hot‑reload, API sur http://localhost:8888/
   ```

6. **Démarrer le frontend en dev** (nouveau terminal) :

   ```bash
   make front            # SvelteKit dev server sur http://localhost:5173/
   ```

7. **Créer un compte** depuis l’UI (Register), puis se connecter.

   * Pour attribuer le rôle **admin** à un utilisateur (facultatif) :

     ```sql
     -- via votre client SQL (psql, TablePlus, etc.)
     UPDATE users SET role = 'admin' WHERE username = 'votre_user';
     ```

---

## Environnements & Ports

* **PostgreSQL** : `5432`
* **API (dev)** : `8888`
* **Frontend (dev)** : `5173`
* **Frontend (Docker prod)** : `3000` (adapter selon votre déploiement)

> Ajustez `PUBLIC_API_BASE_URL` côté frontend selon l’URL réelle de l’API.

---

## Migrations (Goose)

Les fichiers SQL se trouvent dans `migrations/` et couvrent la création des tables :

* `users`, `games`, `game_players`, `game_moves`, `reports`, `push_subscriptions`, etc.

Commandes utiles :

```bash
make migrate-up      # appliquer toutes les migrations
make migrate-down    # rollback (attention aux données)
```

---

## Développement

* **API** : logique métier, validation de coups, calculs de scores, fin de partie, Web Push.
* **Frontend** : Board 15×15, rack avec drag‑and‑drop, aperçu de score en direct (simulate), historique, gestion des reports.

Checklist dev :

* Configuration `.env` OK (API & Frontend).
* CORS : autoriser l’origine du Frontend dans l’API si différent domaine/port.
* Service Worker : nécessaire pour les notifications push (en prod sous HTTPS, ou localhost).

Tests :

```bash
cd api && go test ./...
# (tests front non fournis par défaut)
```

---

## Makefile (principales cibles)

```bash
make db            # lance Postgres via docker-compose
make migrate-up    # applique les migrations
make migrate-down  # revert les migrations
make air           # API en dev (live reload)
make front         # Frontend en dev
```

---

## Déploiement (piste rapide)

* **API** : builder une image à partir de `api/Dockerfile`, fournir `.env` (JWT\_SECRET, POSTGRES\_URL, VAPID\_\*).
* **Frontend** : builder (SvelteKit) puis servir (Node adapter) sur port 3000, derrière un reverse proxy (Nginx/Caddy).
* **Base** : Postgres managé ou conteneur ; sauvegardes via volume.
* **HTTPS** : terminez TLS au niveau du proxy ; exposez l’API sous un domaine et configurez `PUBLIC_API_BASE_URL` côté Frontend.

---

## Dépannage

* **401/403 systématiques** : vérifiez `JWT_SECRET` et l’en‑tête `Authorization: Bearer …` côté Frontend.
* **CORS** : si le Frontend est sur un autre domaine/port, configurez CORS côté API.
* **Migrations échouent** : la chaîne `POSTGRES_URL` est incorrecte ou la DB n’est pas accessible.
* **Push notifications** : assurez la présence du Service Worker et la cohérence des clés VAPID (API/Front).

---

## Licence

Projet éducatif/démo — adaptez la licence à votre contexte. Contributions bienvenues.
