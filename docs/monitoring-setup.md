# Configuring monitoring (Grafana + Prometheus)

## Reqirements

[Install Docker](https://docs.docker.com/engine/install/) using the
appropriate method for your OS.

## Configuration

To begin you need to pull down the repo (or use the zipped source from
the release). I'll use the [release](releases) code for this example.
The `zip` archive is for Windows, the `tar.gz` for everything else.
Unzip the source in a directory of your choice and open a shell/cmd
prompt. At this point if you can not progress without [Docker](https://docs.docker.com/engine/install/)
installed. Go install it if you haven't already. For this example I'll
be running everything in docker -- including the bridge. So type the
following from the root folder to stand up everything:

`docker compose -f docker-compose-all-src.yml up -d --build`

Youll see output about downloading images. After completion, everything
is running successfully in the background.

* spr_bridge is running on port `:5555`
* prometheus is running on port `:9090`
* grafana is running on port `:3000`

You may point your miners the IP address of the computer you installed
on at port `5555`. If you're unsure about your current IP then run
`ipconfig` on Windows and `ip a ls` in Linux. You'll put this IP and
the port into your miner config.

## Accessing Grafana

Assuming the setup went correctly you'll be able to access grafana by
visiting `http://127.0.0.1:3000/d/z73gHk89e1/sprb-monitoring`. Grafana
will asl for login. The default username and password is `admin`.
Grafana will prompt you to change the password but you can just ignore
it (hit skip). You will then be redirected to the mining dashboard.
It'll look empty on fresh installation until you start getting info
from your miners.

At this point you're configured and good to go. Many of the stats on the
graph are averaged over a configurable time period (24hr default - use
the 'resolution' dropdown on the top left of the page to change this),
so keep in mind that the metrics might be incomplete during this initial
period.

Also note that there are 'wallet_filter' and 'show_balances' dropdowns
as well. These filter the database and hide your balance if you don't
want that exposed. The monitoring UI is also accessable on any device on
your local network (including your phone!) if you use the host computers
ip address -- just type in the ip and port such as
`http://192.168.0.25:3000` (this is an example, this exact link
probablly wont work for you).
