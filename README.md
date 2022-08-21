# Go OAuth2 Client

Source code of <https://www.youtube.com/shorts/ychqnQCKCg8>.

It's just a simple demo of how to make an OAuth2 client. In this specific example using GitHub.

Go to <https://github.com/settings/applications/new>
and create a new OAuth application.
Set the redirect callback URI to <http://localhost:3000/api/v1/oauth2/github/callback>.
Copy the client ID and secret and put them in a `.env` file like so:

```env
GITHUB_CLIENT_ID=YOUR_CLIENT_ID_HERE
GITHUB_CLIENT_SECRET=YOUR_CLIENT_SECRET_HERE
```

Build and run:

```bash
go build
./go-oauth2-client-short
```

Visit <http://localhost:3000/api/v1/oauth2/github/redirect>.
