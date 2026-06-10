FROM node:24-alpine
RUN corepack enable
WORKDIR /app

# install dependencies
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .
RUN pnpm install --frozen-lockfile 

# copy all sources
COPY web/ web/
COPY public/ public/
COPY index.html vite.config.ts tsconfig.app.json tsconfig.json tsconfig.node.json .

EXPOSE 5173
CMD ["pnpm", "run", "dev", "--host"]
