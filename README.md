# localtunnel

Expose yourself from NAT or Firewall to the World!

## Features

* Show your work to anyone
* Use the API to test webhooks
* Test your UI in cloud browsers

## How it works

             +--------------------+
             | Localtunnel Server |
             |--------------------|         +----------+
             | Backend | Frontend |<--------+TCP Client|
             +---------+----------+         +----------+
                ^  ^^
                |  ||
                |  ||
         Control|  ||Proxy
      Connection|  ||Connections
                |  ||
             +--+--++-------------+         +----------+
             | Localtunnel Client +-------->|TCP Server|
             +--------------------+         +----------+

In a nutshell, localtunnel consists of two components. A server and a client. The server uses express to listen for incoming requests. These requests can either be from a connecting client that wishes to expose its local port to the public net or a client wishing to connect to an already established service at a subdomain.

When a request comes in from the localtunnel client component, it makes a request to https://thelocaltunnelserver/?new and localtunnel server fires up a new TCP server on a randomly generated port greater than 1023 (non-privileged).

The server then returns this randomly generated port to the localtunnel client and gives the client 5 seconds to connect. If the localtunnel client does not establish a connection to the TCP port within 5 seconds, the server is closed and the localtunnel client will have to reconnect to try again.

If the localtunnel client is able to connect to the localtunnel server’s randomly generated TCP port, by default it opens 10 TCP sockets to the server. These connections are held open, even if no data is being transferred. The localtunnel client then waits for requests to come in over any of these 10 TCP sockets. When a request comes in, it is piped to a TCP client that connects to localhost for the desired service.

In order to expose the localtunnel client’s local service to the web, the localtunnel server waits for requests to come in on the subdomain chosen by the localtunnel client. If it matches the subdomain of a currently connected client, the localtunnel server proxies the request to one (or more) of the 10 TCP sockets being held open by the localtunnel client.

## Building project

Get the package.

```sh
go get -u github.com/bleenco/localtunnel
```

cd into directory.

```sh
cd $GOPATH/src/github.com/bleenco/localtunnel
```

Install tools.

```sh
make get_tools
```

Build the project.

```sh
make build
```

## Docker Image

Change ENV DOMAIN and ENV SECURE variables in `Dockerfile` to fit your needs, then build Docker image.

```sh
make docker_image
```

Run container from `localtunnel` image.

```sh
docker run -dit --restart always -p 1234:1234 --name localtunnel localtunnel
```

## Usage Example

Start the server on port 1234 hosting on domain example.com.

```sh
./lt-server -p 1234 -d example.com
```

Create the tunnel to localhost:8000.

```sh
./lt-client -h http://example.com:1234 -p 8000
```

You should get generated URL, something like `7e400f6d.example.com`.

Open your browser on `http://7e400f6d.example.com:1234` to check if its working (it should :) then share the URL with your friends.

## Licence

MIT
