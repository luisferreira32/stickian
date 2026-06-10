FROM node:24-alpine
RUN corepack enable
WORKDIR /app

# install dependencies
<<<<<<< HEAD
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .
RUN pnpm install --frozen-lockfile 
=======
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install
>>>>>>> 722c19a (trying to fix pnpm build error)

# copy all sources
COPY web/ web/
COPY public/ public/
COPY index.html vite.config.ts tsconfig.app.json tsconfig.json tsconfig.node.json ./

EXPOSE 5173
CMD ["pnpm", "run", "dev", "--host"]
