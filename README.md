# MQTT5 Explorer (Go)

[![Issues](https://img.shields.io/github/issues/Omniaevo/mqtt5-explorer-go)](https://github.com/Omniaevo/mqtt5-explorer-go/issues) [![Workflow status](https://img.shields.io/github/actions/workflow/status/Omniaevo/mqtt5-explorer-go/wails.yml)](https://github.com/Omniaevo/mqtt5-explorer-go/actions) ![Last commit](https://img.shields.io/github/last-commit/Omniaevo/mqtt5-explorer-go) [![Latest release](https://img.shields.io/github/v/release/Omniaevo/mqtt5-explorer-go)](https://github.com/Omniaevo/mqtt5-explorer-go/releases) [![License](https://img.shields.io/github/license/Omniaevo/mqtt5-explorer-go)](https://github.com/Omniaevo/mqtt5-explorer-go/blob/master/LICENSE)

## About this project

The aim of this project is to bring the users a client app capable of making use of all the features of the version 5 of the MQTT protocol. The lack of any application that can offer the compatibility with the newer version of the protocol forced us to implement one to test the data of MQTT brokers workwise, why not to share this tool with others that may have the same issue?

## Screenshots

### Dark theme

![Client screenshot (dark theme)](screenshots/client-connection.png)

### Light theme

![Client screenshot (light theme)](screenshots/client-connection-white.png)

## Project setup

You can configure the project by editing `wails.json`. More information about the project settings can be found here: [https://wails.io/docs/reference/project-config](https://wails.io/docs/reference/project-config)

### Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser and have access to your Go methods, there is also a dev server that runs on [http://localhost:34115](http://localhost:34115). Connect to this in your browser, and you can call your Go code from devtools.

## Compiles for production

To build a redistributable, production mode package, use `wails build`.

## Get involved

See:

- [The code of conduct](CODE_OF_CONDUCT.md).
- [The contribution guidelines](.github/contributing.md).

## Contributors

[![Contributors](https://contrib.rocks/image?repo=Omniaevo/mqtt5-explorer-go)](https://github.com/Omniaevo/mqtt5-explorer-go/graphs/contributors)
