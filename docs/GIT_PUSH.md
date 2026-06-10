# Push to GitHub

Create an empty private repository, for example:

```text
git@github.com:vorkauteer/ITSProto-Windows.git
```

Then push:

```bash
git init
git branch -M main
git remote add origin git@github.com:vorkauteer/ITSProto-Windows.git

git add .
git commit -m "v0.1: add minimal Windows GUI client"
git push -u origin main

git tag -a v0.1-gui-launcher -m "Minimal Windows GUI launcher"
git push origin v0.1-gui-launcher
```
