# AGENTS.md

This file provides guidance to coding agents when working with code in this repository.

## Project Overview

FES (Festival Edit System) is a Japanese CMS for managing festival/event websites.

## 構成

| ディレクトリ | 内容 | 詳細 |
|---|---|---|
| `app/` | Go バックエンド（公開サイト・管理 API・アーカイブ）。Google App Engine にデプロイ | [app/AGENTS.md](app/AGENTS.md) |
| `manage/` | 管理画面 SPA（React + Vite + MUI）。ビルド成果物を `app/` に embed して `/manage/` で配信 | [manage/AGENTS.md](manage/AGENTS.md) |
| `maps/` | 会場マップの React SPA。`app/` とは独立して開発し、ビルド成果物を `app/` に embed して `/maps/` で配信 | [maps/AGENTS.md](maps/AGENTS.md) |

各ディレクトリで作業するときは、そのディレクトリの `AGENTS.md` を参照すること。

## 作業手順（docs/skills）

決まった手順の作業は `docs/skills/<名前>/SKILL.md` にまとめている。該当する作業のときは先に読むこと。

| スキル | 使う場面 |
|---|---|
| [fes-setup](docs/skills/fes-setup/SKILL.md) | 環境の初期セットアップ、ローカル起動、ログイン・管理者登録 |
| [fes-deploy](docs/skills/fes-deploy/SKILL.md) | manage / maps の embed 反映と GAE への本番デプロイ |
| [fes-archive](docs/skills/fes-archive/SKILL.md) | 静的アーカイブの生成・GCS アップロード・公開 |
| [manage-feature](docs/skills/manage-feature/SKILL.md) | 管理画面への API・画面の追加 |
| [maps-update](docs/skills/maps-update/SKILL.md) | 会場マップの店舗・画像・GPS 基準点の更新 |

## 構想・方針（docs/plans）

まだ実装していない方針や構想は `docs/plans/` に置いている。関連する作業の前に目を通すこと。

| 文書 | 内容 |
|---|---|
| [maps-automation](docs/plans/maps-automation.md) | 会場マップのデータ作成の半自動化（現状と改善案） |
| [event-master](docs/plans/event-master.md) | 出店者・アーティストのマスタ化と、店主による店舗情報の編集 |

## Key Conventions

- Management UI, comments, and commit messages are in Japanese.
- No test files exist in the repo; manual/integration testing is the current approach.
- `manage/` と `maps/` の ESLint は同じ方針（ESLint 10 flat config、`eslint-plugin-react` は不採用）。詳細は各 `AGENTS.md` を参照。
