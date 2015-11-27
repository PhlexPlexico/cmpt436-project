# secure-chat

This is intended to be the final implementation for the project.

### Setup

Setup is a bit more involved for this prototype. Make sure you have [Bower](http://bower.io/) and [Go](https://golang.org/) installed, and then run:

`bower install`

in the `app/` directory to get all of the website dependencies. After that, you will need to create a certificate and key (named `cert.crt` and `key.key`) for https and wss connection, and place them in the `main/` directory. Specifically, just execute the following command in the 'main/' directory:

`sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout key.key -out cert.crt`

The program expects to find a config file containing the session secret, all client ids and client secrets, and all url and port information, in a specifically formatted json file called '.config.json', within 'main/'. You probably need to get this from William. 

Within that file, you should edit the website_url variable to be the url on which you will be hosting the server. You should also edit https_portnum and http_portnum. 

website_url should be formatted like:

  "https://localhost"
  
https_portnum and http_portnum should be formatted like:

  ":XXXX"

(no trailing slashes).

After that, run

`go run main.go`

in the `main/` directory, and the site will be accessible via the url and portnum you specified in 
.config.json

