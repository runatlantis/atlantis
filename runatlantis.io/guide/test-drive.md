# Test Drive
To test drive Atlantis on an example repo, download the latest release for your architecture:
[https://github.com/runatlantis/atlantis/releases](https://github.com/runatlantis/atlantis/releases)

Once you've extracted the archive, run:
```bash
./atlantis testdrive
```

This mode sets up Atlantis on a test repo so you can try it out. It will
- fork an example terraform project into your GitHub account
- install terraform (if not already in your PATH)
- install ngrok so we can expose Atlantis to GitHub
- start Atlantis so you can execute commands on the pull request

## Next Steps

When you're ready to try out Atlantis on your own repos then read [Getting Started](getting-started.html).
