FROM node:12.9.1-alpine as react-build
WORKDIR /app
COPY ./package.json ./package-lock.json ./
RUN npm install
COPY . ./
RUN npm run build

FROM nginx:1.17-alpine
COPY --from=react-build /app/build /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]