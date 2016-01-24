# sharepass - a simple tool for sharing passwords over a local network

## How to use:

_tl;dr_ It opens a temporary server to send a password once and copies the URL automatically.

1. Install it with `GO15VENDOREXPERIMENT=1 go get github.com/gophergala2016/sharepass`
2. Run `sharepass`
3. A temporary local URL will be copied to your clipboard automatically, and the server will fork into the background and continue running.
4. Send this URL to the recipient via whatever channel you want. The server will automatically exit when they successfully load the page.
5. They copy the password from the page and use it.
6. The server will automatically exit after a configurable timeout.

## FAQ

- Isn't this insecure?

  - Yes, but it's better than sending the password directly in plaintext and having it be permanently stored on somebody else's server.

- Shouldn't you use PGP / other existing secure software for this?

  - Probably. This is a compromise designed for ease-of-use. The main benefit is that the recipient doesn't need to have anything installed or any understanding of information security or cryptography.
