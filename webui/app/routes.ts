import { type RouteConfig, index, route } from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  route("login", "routes/login.tsx"),
  route("dashboard", "routes/dashboard.tsx"),
  route("proxy-hosts", "routes/proxy-hosts.tsx"),
  route("certificates", "routes/certificates.tsx"),
  route("monitoring", "routes/monitoring.tsx"),
  route("nginx-configs", "routes/nginx-configs.tsx"),
  route("nginx-configs/new", "routes/nginx-configs/new.tsx"),
  route("nginx-configs/:id", "routes/nginx-configs/edit.tsx"),
  route("nginx-templates", "routes/nginx-templates.tsx"),
  route("nginx-templates/new", "routes/nginx-templates/new.tsx"),
  route("nginx-templates/:id", "routes/nginx-templates/edit.tsx"),
] satisfies RouteConfig;
