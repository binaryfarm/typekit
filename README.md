# TypeKit #

![golang](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white) ![typescript](https://img.shields.io/badge/TypeScript-007ACC?style=for-the-badge&logo=typescript&logoColor=white)  ![apache-2.0](https://img.shields.io/badge/Apache--2.0-green?style=for-the-badge)


### Sumamry ###

TypeKit is a typescript runtime written in Golang. This was based on [goja](https://github.com/dop251/goja) and [sobek](https://github.com/grafana/sobek) which is a fork of goja. We also had forked [goja](https://github.com/dop251/goja) to add improvements we internally called `bagel` and so we've merged our work with upstream sobek changes and threw on [esbuild](https://github.com/evanw/esbuild) to transpile. While not as fast as Node(V8)/Deno(V8)/Bun(JSC) it's written in *pure go*.

## WARNING ##
This is in no way shape or form ready for use outside of tinkering/exploring. We welcome contributions to help improve the runtime but until [typescript-go](https://github.com/microsoft/typescript-go) is finished, we'll have to transpile. Which is fine enough for now.


### Roadmap ###
- Continue developing the runtime to be in partiy with latest ECMAScript standards.
- Utilize typescript-go once finished to compile typescript into javascript
- Run javascript on runtime and expose golang features like `chan` to improve concurrency.
- Have fun along the way.


### What this is NOT ###
It's not an attempt to shatter any records or provide world-class native typescript binaries (maybe someday).
