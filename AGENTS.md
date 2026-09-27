# AGENTS.md

This file provides guidance to coding agents when working with code in this repository.

## Project Overview

FES (Festival Edit System) is a Japanese CMS for managing festival/event websites.

## 構成

| ディレクトリ | 内容 | 詳細 |
|---|---|---|
| `app/` | Go バックエンド（公開サイト・管理 API・アーカイブ）。Google App Engine にデプロイ | [app/AGENTS.md](app/AGENTS.md) |
| `manage/` | 管理画面 SPA（React + Vite + MUI）。ビルド成果物を `app/` に embed して `/manage/` で配信 | [manage/AGENTS.md](manage/AGENTS.md) |
| `maps/` | 会場マップの React SPA。`app/` とは独立して開発・デプロイする | [maps/AGENTS.md](maps/AGENTS.md) |

各ディレクトリで作業するときは、そのディレクトリの `AGENTS.md` を参照すること。

## Key Conventions

- Management UI, comments, and commit messages are in Japanese.
- No test files exist in the repo; manual/integration testing is the current approach.
- `manage/` と `maps/` の ESLint は同じ方針（ESLint 10 flat config、`eslint-plugin-react` は不採用）。詳細は各 `AGENTS.md` を参照。
