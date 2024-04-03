# Test Drive
To test drive Atlantis on an example repo, download the latest release from
[GitHub](https://github.com/runatlantis/atlantis/releases)

Once you've extracted the archive, run:
```bash
./atlantis testdrive
```

This mode sets up Atlantis on a test repo so you can try it out. It will
- Fork an example Terraform project into your GitHub account
- Install Terraform (if not already in your PATH)
- Install [ngrok](https://ngrok.com/) so we can expose Atlantis to GitHub
- Start Atlantis so you can execute commands on the pull request

## Next Steps
* If you're ready to test out running Atlantis on **your repos** then read [Testing Locally](testing-locally.md).
* If you're ready to properly install Atlantis on real infrastructure then head over to the [Installation Guide](../docs/installation-guide.md).
