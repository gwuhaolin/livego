# LiveGo Web UI

Access to Livego Web UI, ex: http://localhost:8090/dashboard


## How to build (for backend developer)

Use the make file :

```shell
make build           # Generate Docker image
make generate-webui  # Generate static contents in `livego/static/` folder.
```

## How to build (only for frontend developer)

- prerequisite: [Node 12.11+](https://nodejs.org) [Npm](https://www.npmjs.com/)

- Go to the directory `webui`

- To install dependencies, execute the following commands:

  - `npm install`

- Build static Web UI, execute the following command:

  - `npm run build`

- Static contents are build in the directory `build`

**Don't change manually the files in the directory `build`**

- The build allow to:
  - optimize all JavaScript
  - optimize all CSS
  - add vendor prefixes to CSS (cross-bowser support)
  - add a hash in the file names to prevent browser cache problems
  - all images will be optimized at build
  - bundle JavaScript in one file

## How to edit (only for frontend developer)

**Don't change manually the files in the directory `build`**

- Go to the directory `webui`
- Edit files in `webui/src`
- Run in development mode :
  - `npm run dev`

## Libraries

- [Node](https://nodejs.org)
- [Npm](https://www.npmjs.com/)
- [React](https://reactjs.org/)