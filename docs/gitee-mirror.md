# Mirroring gr33n to Gitee (GitHub stays source of truth)

Goal: Chinese forums / Gitee discoverability, while you keep developing on GitHub (`dgang0404/gr33n`).

## Reality check (tokens)

| Approach | Auto on GitHub push? | What you need |
|----------|----------------------|---------------|
| **A. Gitee Pull mirror** (recommended first) | Yes, if you enable auto + webhook | **GitHub** PAT pasted into Gitee (not a Gitee token). Gitee pulls *from* GitHub. |
| **B. GitHub Action → push to Gitee** | Yes | **Gitee** PAT (and often SSH key) stored as GitHub Actions secrets |
| **C. Two remotes, manual** | No | `git push github` and `git push gitee` when you remember |
| **D. Weekly cron only** | No (scheduled) | Same as A (manual sync button) or B with `cron:` |

There is **no** magic zero-credential auto-sync. Someone has to authenticate either the pull (GitHub→Gitee) or the push (Gitee←GitHub). Public clone URLs alone don't get webhooks or scheduled pulls on Gitee's free tier without that setup.

## Path A — Gitee pulls GitHub (best match for "I push GitHub, Gitee catches up")

1. Create a repo on [gitee.com](https://gitee.com) (e.g. `gr33n`, public, empty — no README if you want a clean mirror).
2. Repo → **管理** → **仓库镜像管理** / mirror settings.
3. Add **Pull** direction: source `https://github.com/dgang0404/gr33n.git`.
4. Prefer **自动从 GitHub 同步** (auto sync). Gitee will ask for a **GitHub personal access token** (`repo` + `admin:repo_hook` for webhook). Create that at GitHub → Settings → Developer settings → Tokens.
5. After the first sync, Gitee should track branches/tags/commits. Pushing to GitHub then triggers (or you click **同步**) Gitee update.

Docs: [Gitee ↔ GitHub mirror help](https://help.gitee.com/repository/settings/sync-between-gitee-github).

Weekly-only: skip auto webhook; open Gitee once a week and hit sync — no Action, no Gitee token.

## Path B — GitHub Action pushes to Gitee (optional later)

Use when Path A is flaky or you want sync on every `main` push from CI.

1. Gitee: create empty `gr33n` repo; create a Gitee private token with repo write.
2. GitHub repo secrets (example names):
   - `GITEE_TOKEN`
   - `GITEE_USERNAME` (your Gitee login)
3. Add a workflow under `.github/workflows/sync-to-gitee.yml` (hub-mirror-action or a plain `git push --mirror`). **Do not commit secrets.** Ask in chat when you're ready and we can drop in a minimal workflow.

## Path C — Local dual remote (one-off / backup)

```bash
# after the Gitee repo exists
git remote add gitee https://gitee.com/<YOUR_GITEE_USER>/gr33n.git
git remote -v
git push gitee main
```

GitHub stays `origin`. Push Gitee when you care about the China-facing copy.

## Chinese forum / Dev.to post

Your unpublished Dev.to draft (`gr33n-devto-post.pdf`) is a strong spine: extraction vs commons, AGPL, self-hosted stack, Guardian on-LAN, invite forks. For Chinese forums:

- Lead with the same thesis (数据不出农场 / AGPL / 拒绝云锁定).
- Link **both** GitHub and Gitee once the mirror exists.
- Update phase counts — the PDF still says “29 phases / Phase 30–31”; current arc is through **211.x** (natural farming + Guardian QA). Refresh numbers before posting.
- Keep the tone peer-to-peer (same as the English draft): builders, not a sales pitch.

When you have a Gitee username + whether you prefer Path A or B, we can wire the remaining steps (remote URL, or the Actions workflow file).
