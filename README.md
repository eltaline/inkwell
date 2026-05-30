# inkwell

`inkwell` is a self-hosted, multi-user knowledge base in the spirit of Obsidian — but it runs on your own server and your team works in it together. Notes are plain Markdown files stored in a Git repository, so every edit is a commit: your data stays portable, diffable, and yours.

It pairs the things people love about Obsidian — wikilinks, backlinks, a graph of how ideas connect, tags, instant search — with the things you actually need to run it for a team: accounts, permissions, concurrent editing, and a full version history backed by Git.

## Features

- **Markdown over Git** — every note is a `.md` file in a Git repo; each save is a commit with an author. Nothing is locked in a proprietary database.
- **Wikilinks & embeds** — link notes with `[[note]]`, alias with `[[note|label]]`, and transclude content with `![[note]]`.
- **Backlinks & graph** — see every note that references the current one (with context), surface unlinked mentions, and explore connections in a local or whole-vault graph view.
- **Tags** — inline `#tags` and front-matter tags, nested tag trees, and tag-based filtering.
- **Full-text search** — fast search across titles, bodies, and tags with a query syntax (`tag:`, `path:`, `title:`) and a Ctrl/Cmd-K quick switcher.
- **Version history** — per-note history, side-by-side diffs, one-click restore, and a vault-wide activity feed of who changed what.
- **Built for teams** — accounts, admin/editor/viewer roles, per-folder access control, concurrent-edit detection with conflict resolution, and live presence.
- **Export & publish** — render notes to PDF or HTML, bulk-export a vault, publish selected notes as read-only share links, or generate a static digital-garden site.

## Quick start

```sh
# with Docker
docker run -d \
  -p 8080:8080 \
  -v /srv/inkwell/data:/data \
  ghcr.io/eltaline/inkwell:latest

# or docker-compose
curl -L https://raw.githubusercontent.com/eltaline/inkwell/main/docker-compose.yml -o docker-compose.yml
docker compose up -d
```

Open `http://localhost:8080`, create the first admin account, and point it at an existing Git repo or let it initialize a fresh vault.

## How it works

A **vault** is a Git repository on disk. Notes are Markdown files with optional YAML front-matter:

```markdown
---
id: 01H...
title: Designing the link index
tags: [architecture, search]
---

The index maps each note to its outbound `[[links]]`, which powers
both [[Backlinks]] and the [[Graph view]].
```

`inkwell` parses links and tags to build the link graph and search index, commits every change back to the repo, and serves it all through a web UI. Because the source of truth is just Markdown in Git, you can still clone the repo, edit in your favorite editor, and push — `inkwell` picks up the changes.

## Status

Under active development. Expect breaking changes before `v1.0`.

## License

MIT
