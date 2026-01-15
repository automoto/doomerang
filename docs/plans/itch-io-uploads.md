## Goal
Upload Doomerang to itch.io with web playable version and downloadable binaries for all platforms.

## Step 1: Create an itch.io Account
If you don't have one, sign up at [itch.io](https://itch.io)

## Step 2: Create a New Game Project

1. Click the **arrow beside your username** (top-right) → **"Upload new project"**
   - Or: Dashboard → **"Create new project"**

2. Fill out the form:

| Field | Value |
|-------|-------|
| **Title** | Doomerang |
| **Project URL** | `your-username.itch.io/doomerang` (auto-generated) |
| **Classification** | Games |
| **Kind of project** | HTML (for web playable) + Downloadable |
| **Release status** | In development / Released |
| **Pricing** | Free / Paid / Name your price |
| **Description** | Your game description |

3. **Cover Image**: Upload a 630x500 image (or 315x250 minimum)

## Step 3: Upload Your Builds

### Option A: Manual Upload (Web UI)
1. Click **"Upload files"**
2. Upload each platform build as a ZIP:
   - `doomerang-windows.zip` → tag as **Windows**
   - `doomerang-mac.zip` → tag as **macOS**
   - `doomerang-linux.zip` → tag as **Linux**
   - `doomerang-web.zip` → check **"This file will be played in the browser"**

### Option B: Butler (Recommended)
Using your existing Makefile targets:

```bash
# First time: login to butler
butler login

# Update ITCH_USER and ITCH_GAME in Makefile, then:
make deploy-all
```

Butler auto-tags platforms based on channel names (`windows`, `mac`, `linux`, `web`).

## Step 4: Configure Web Embed

For the web build to be playable in browser:

1. After uploading, find your web build in the uploads list
2. Check **"This file will be played in the browser"**
3. Set **Viewport dimensions** (e.g., 800x600 or your game resolution)
4. Choose embed type:
   - **"Embed in page"** - plays directly on page
   - **"Click to launch fullscreen"** - better for games

## Step 5: Web Build Requirements

Your `dist/web/` folder needs:
```
index.html      ← Entry point (required name)
doomerang.wasm
wasm_exec.js
```

**Important for WASM:**
- All paths must be **relative** (not absolute)
- Filenames are **case-sensitive**
- Max 500MB total, max 1000 files

## Step 6: Set Visibility & Publish

1. **Draft** (default) - Only you can see it, good for testing
2. **Restricted** - Share with testers via access keys
3. **Public** - Visible to everyone, appears in search

When ready: Edit project → Change visibility to **Public** → Save

---

## Quick Checklist

- [ ] Create itch.io account
- [ ] Create new project with title "Doomerang"
- [ ] Upload cover image (630x500)
- [ ] Add screenshots (3-5 recommended)
- [ ] Run `make build-all` to create all platform builds
- [ ] Run `butler login` (first time only)
- [ ] Update `ITCH_USER` in Makefile
- [ ] Run `make deploy-all` to upload all builds
- [ ] Enable web embed for the web build
- [ ] Set viewport dimensions
- [ ] Add description and tags
- [ ] Set visibility to Public

---

## References
- [Your first itch.io page](https://itch.io/docs/creators/getting-started)
- [Uploading HTML5 games](https://itch.io/docs/creators/html5)
- [Creator FAQ](https://itch.io/docs/creators/faq)
