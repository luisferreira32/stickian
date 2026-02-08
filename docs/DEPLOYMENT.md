# Deployment

Internal documentation for deployment of the project.

## Pre-requesites

- go
- pnpm
- brotli
- tar (\*)
- gzip (\*)

(\*) you can use other tools copy over build artifacts

## Step by step

To deploy the application into a remote host of your choice:

1. Bundle the React application

```bash
pnpm build && brotli dist/assets/*.js dist/assets/*.css
```

2. Compile the Go server

```bash
go build -o stickian-server ./server/
```

3. Bundle everything to transfer to the host

```bash
tar czvf bundle.tar.gz stickian-server dist/
```

4. Transfer to the host

This step depends where you want to run the project. If you just want to test it locally, no need to bundle it, and just run the `stickian-server` right away.

5. Run it on the host

```bash
tar xvf bundle.tar.gz && ./stickian-server
```
