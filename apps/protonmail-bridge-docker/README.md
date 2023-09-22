# ProtonMail IMAP/SMTP Bridge Docker Container

![build badge](https://github.com/larivierec/protonmail-bridge-docker/workflows/release/badge.svg)

This is an unofficial Docker container of the [ProtonMail Bridge](https://protonmail.com/bridge/). Some of the scripts are based on [Hendrik Meyer's work](https://gitlab.com/T4cC0re/protonmail-bridge-docker).


GitHub: [https://github.com/larivierec/protonmail-bridge-docker](https://github.com/larivierec/protonmail-bridge-docker)

## ARM Support

We now support ARM devices (`arm64` )! Use the images tagged with `build`. See next section for details.

## Tags

There are two types of images.
 - `deb`: Images based on the official [.deb release](https://protonmail.com/bridge/install). It only supports the `amd64` architecture.
 - `build`: Images based on the [source code](https://github.com/ProtonMail/proton-bridge). It supports `amd64`, `arm64`, and `arm/v7`. Supporting to more architectures is possible. PRs are welcome.

tag | description
 -- | --
`latest` | latest `deb` image
`[version]` | `deb` images
`build` | latest `build` image
`[version]-build` | `build` images

## Initialization

To initialize and add account to the bridge, run the following command.

```
docker run --rm -it -v protonmail:/root larivierec/protonmail-bridge init
```

Wait for the bridge to startup, use `login` command and follow the instructions to add your account into the bridge. Then use `info` to see the configuration information (username and password). After that, use `exit` to exit the bridge. You may need `CTRL+C` to exit the docker entirely.

## Run

To run the container, use the following command.

```
docker run -d --name=protonmail-bridge -v protonmail:/root -p 1025:25/tcp -p 1143:143/tcp --restart=unless-stopped larivierec/protonmail-bridge
```

## Security

Please be aware that running the command above will expose your bridge to the network. Remember to use firewall if you are going to run this in an untrusted network or on a machine that has public IP address. You can also use the following command to publish the port to only localhost, which is the same behavior as the official bridge package.

```
docker run -d --name=protonmail-bridge -v protonmail:/root -p 127.0.0.1:1025:25/tcp -p 127.0.0.1:1143:143/tcp --restart=unless-stopped larivierec/protonmail-bridge
```

Besides, you can publish only port 25 (SMTP) if you don't need to receive any email (e.g. as a email notification service).

## Compatibility

The bridge currently only supports some of the email clients. More details can be found on the official website. I've tested this on a Synology DiskStation and it runs well. However, you may need ssh onto it to run the interactive docker command to add your account. The main reason of using this instead of environment variables is that it seems to be the best way to support two-factor authentication.

## Bridge CLI Guide

The initialization step exposes the bridge CLI so you can do things like switch between combined and split mode, change proxy, etc. The [official guide](https://protonmail.com/support/knowledge-base/bridge-cli-guide/) gives more information on to use the CLI.

## Build

For anyone who want to build this container on your own (for development or security concerns), here is the guide to do so. First, you need to `cd` into the directory (`build`, depending on which type of image you want). Then just run the docker build command
```
docker build .
```

That's it. The `Dockerfile` and bash scripts handle all the downloading, building, and packing. You can also add tags, push to your favorite docker registry, or use `buildx` to build multi architecture images.

## Credit

Initial credit goes to github user: `shenxn`
This build is based off the source version rather than the `.deb` package which seems to be the maintained version from that repository.