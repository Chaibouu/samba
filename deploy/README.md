# Déployer Posta sur un VPS

Ce dossier contient tout ce qu'il faut pour faire tourner Posta en production sur
un seul serveur, et l'utiliser pour envoyer les emails de tes projets Node.

```
Tes projets ──API HTTPS──▶ Posta (ce VPS) ──SMTP──▶ Fournisseur (Brevo, SES…) ──▶ Destinataires
```

Posta ne livre pas les emails lui-même : il passe par un fournisseur SMTP. Il gère
tout le reste : API, templates, file d'attente, retries, suivi, bounces, logs.

| Fichier | Rôle |
|---|---|
| `compose.yml` | Posta + worker + PostgreSQL + Redis + Caddy (HTTPS automatique) |
| `Caddyfile` | Reverse proxy HTTPS devant Posta |
| `.env.example` | Les secrets à remplir |
| `node/posta.mjs` | Client Node minimal, sans dépendance |

---

## 1. Choisir un fournisseur SMTP

| Fournisseur | Gratuit | Pour qui |
|---|---|---|
| **Brevo** | 300 emails/jour | Le plus simple pour démarrer. Recommandé. |
| **Amazon SES** | — (~0,10 $ / 1000 emails) | Gros volume, configuration plus technique |
| **Mailjet** | 200 emails/jour | Alternative à Brevo |

Crée un compte, puis récupère les **identifiants SMTP** : hôte, port 587,
utilisateur, mot de passe. Chez Brevo : *Paramètres → SMTP & API*.

## 2. Préparer le VPS

N'importe quel VPS Linux avec 2 Go de RAM convient (Hetzner, OVH, Contabo…).

```bash
# Docker
curl -fsSL https://get.docker.com | sh

# Pare-feu : seuls SSH, HTTP et HTTPS sont ouverts
ufw allow OpenSSH && ufw allow 80 && ufw allow 443 && ufw enable
```

Dans la zone DNS de ton domaine, crée un enregistrement **A** :
`posta.tondomaine.com → IP du VPS`. Attends qu'il soit actif (`ping posta.tondomaine.com`)
avant de démarrer, sinon Caddy ne pourra pas obtenir le certificat HTTPS.

## 3. Installer et démarrer

```bash
git clone https://github.com/Chaibouu/samba.git posta
cd posta
git checkout fix/dashboard-xss-and-query-token   # ou main une fois la PR fusionnée
cd deploy

cp .env.example .env
nano .env
```

Remplis `.env`. Pour chaque secret, génère une valeur avec :

```bash
openssl rand -hex 32
```

Puis :

```bash
docker compose up -d --build
docker compose ps                # attendre que posta soit « healthy »
```

Le premier build prend quelques minutes (dashboard + binaire Go).
Ouvre ensuite `https://posta.tondomaine.com` et connecte-toi avec
`POSTA_ADMIN_EMAIL` / `POSTA_ADMIN_PASSWORD`.

> ⚠️ Garde une copie de `POSTA_ENCRYPTION_KEY` en lieu sûr. Si tu la perds, les
> mots de passe SMTP enregistrés dans Posta deviennent illisibles.

## 4. Configurer dans le dashboard

1. **Serveur SMTP** : ajoute les identifiants de ton fournisseur (port 587, STARTTLS).
2. **Domaine d'envoi** : ajoute le domaine depuis lequel tu envoies
   (ex. `tondomaine.com`) et crée les enregistrements DNS que Posta affiche.
   Ajoute aussi ceux que ton fournisseur SMTP te donne (SPF / DKIM).
   Sans SPF, DKIM et DMARC, tes emails finiront en spam.
3. **Un workspace par projet** : chaque projet a ainsi ses propres logs,
   templates et statistiques.
4. **Clé API** : dans chaque workspace, crée une clé avec le scope `send`.
   Elle commence par `psk_` et n'est affichée qu'une fois.

## 5. Envoyer depuis un projet Node

Dans le `.env` du projet :

```bash
POSTA_URL=https://posta.tondomaine.com
POSTA_API_KEY=psk_xxxxxxxx
```

Copie `node/posta.mjs` dans le projet, puis :

```js
import { sendEmail, sendTemplate } from "./posta.mjs";

await sendEmail({
  from: "Mon App <noreply@tondomaine.com>",
  to: ["client@example.com"],
  subject: "Bienvenue !",
  html: "<h1>Bienvenue</h1>",
});

// Avec un template créé dans le dashboard ({{ name }} dans le HTML)
await sendTemplate({
  template: "welcome",
  to: ["client@example.com"],
  language: "fr",
  template_data: { name: "Awa" },
});
```

L'appel renvoie `{ id, status: "queued" }` : l'email est mis en file d'attente,
puis le worker l'envoie. Son statut s'affiche dans le dashboard.

Pour tester rapidement depuis ton poste :

```bash
POSTA_URL=https://posta.tondomaine.com POSTA_API_KEY=psk_... \
POSTA_FROM="Test <noreply@tondomaine.com>" node node/posta.mjs toi@example.com
```

Il existe aussi un SDK officiel : <https://github.com/goposta/posta-node>.

> La clé API ne doit **jamais** se retrouver dans du code exécuté par le navigateur.
> Appelle Posta uniquement depuis ton backend (route API Next.js, serveur Express…).

---

## Exploitation

```bash
docker compose ps                         # état des services
docker compose logs -f worker             # suivre les envois
docker compose pull && docker compose up -d --build   # mise à jour (après git pull)
```

**Sauvegarde de la base** (à planifier avec cron) :

```bash
docker compose exec -T posta-db pg_dump -U posta posta | gzip > posta-$(date +%F).sql.gz
```

**Emails système** (mot de passe oublié, invitations) : optionnel. Ajoute les
variables `POSTA_SYSTEM_SMTP_*` (voir `../.env.example`) dans le bloc
`x-posta-env` de `compose.yml`.
