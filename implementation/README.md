# implementation

This is intended to be the final implementation for the project.

### Setup

Make sure you have [Bower](http://bower.io/) and [Go](https://golang.org/) installed, and then run:

`bower install`

in the `app/` directory to get all of the website dependencies. You'll then need to add the following code inside of `app/bower_components/iron-a11y-keys-behavior/iron-a11y-keys-behavior.html` at line `253` to fix an outstanding bug:

```
  ready: function() {
    this.keyEventTarget = document.body;
  },
```

After that, you will need to create a certificate and key (named `cert.crt` and `key.key`) for https and wss connection, and place them in the `main/` directory. Specifically, just execute the following command in the 'main/' directory:

`openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout key.key -out cert.crt`

The program expects to find a config file containing the session secret, all client ids and client secrets, and all url and port information, in a specifically formatted json file called '.config.json', within 'main/'. You probably need to get this from William. 

Within that file, you should edit the website_url variable to be the url on which you will be hosting the server. You should also edit https_portnum and http_portnum. 

website_url should be formatted like:

  "localhost" or "www.williamvanderkamp.com"

(no https:// prefix).

https_portnum and http_portnum should be formatted like:

  ":XXXX"

(no trailing slashes).

After that, run

`go get`

in the `main/` directory. It'll tell you it can't load local imports in non-local package, but it should install all the relevant remote packages. Next, run

`go run main.go`

in the `main/` directory, and the site will be accessible via the url and portnum you specified in .config.json.
