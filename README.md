# Cloudflare Dynamic DNS

Script written in Go to update DNS Record's IP address. <br />
Helpful for folks wanting to expose their home servers to the internet without static IP via Cloudflare DNS. <br />

## Usage

- Build Docker image. <br />

```
docker build --tag cloudflare-ddns .
```

- Start a container with required `env` variables.
  - NOTE: For better security pass an env file using `--env-file` flag in the docker run command.

```
CLOUDFLARE_API_TOKEN=xxxxxxxx \
ZONE_ID=xxxxxxxx \
docker run -d --name cloudflare-ddns \
-e CLOUDFLARE_API_TOKEN \
-e ZONE_ID \
-e RECORD_NAME=xxxxxxxx \
--restart unless-stopped \
cloudflare-ddns
```
