---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

pageClass: home-custom

hero:
  name: Atlantis
  text: Automatización de pull requests para Terraform
  tagline: Ejecuta workflows de Terraform con facilidad
  image: /hero.png
  actions:
    - theme: brand
      text: Empezar
      link: /guide
    - theme: alt
      text: ¿Qué es Atlantis?
      link: /blog/2017/introducing-atlantis
    - theme: alt
      text: Únete en Slack
      link: https://slack.cncf.io/

features:
  - title: Menos errores
    details: "Detecta errores en la salida de terraform plan antes de aplicar los cambios. Asegura que los cambios se apliquen antes de hacer merge."
    icon: ✅
  - title: Autonomía para los desarrolladores
    details: "Los desarrolladores pueden enviar pull requests de Terraform de forma segura sin credenciales. Exige aprobaciones para los applies."
    icon: 💻
  - title: Registros de auditoría inmediatos
    details: "Logs detallados de los cambios de infraestructura, las aprobaciones y las acciones de los usuarios. Configura aprobaciones para los cambios en producción."
    icon: 📋
  - title: Probado a gran escala
    details: "Usado por empresas líderes para gestionar más de 600 repositorios con 300 desarrolladores. En producción desde 2017."
    icon: 🌍
  - title: Autoalojado
    details: "Tus credenciales siguen siendo tuyas. Desplegable en VMs, Kubernetes, Fargate, etc. Compatible con GitHub, GitLab, Bitbucket y Azure DevOps."
    icon: ⚙️
  - title: Código abierto
    details: "Atlantis es un proyecto de código abierto con una comunidad sólida, impulsado por contribuciones voluntarias."
    icon: 🌐

---
