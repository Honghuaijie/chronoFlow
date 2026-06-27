FROM node:20-alpine AS builder

WORKDIR /src
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:1.27-alpine

COPY nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=builder /src/dist /usr/share/nginx/html

EXPOSE 80
