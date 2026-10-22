---
title: Resultados de la encuesta de usuarios de Atlantis
lang: en-US
---

# Resultados de la encuesta de usuarios de Atlantis

En abril de 2024, el Core Atlantis Team lanzó una encuesta anónima de nuestros usuarios. Durante los dos meses en que la encuesta estuvo abierta recibimos 354 respuestas, que usaremos para comprender mejor las necesidades de nuestra comunidad y ayudar a priorizar nuestra hoja de ruta.

En general, los resultados a continuación muestran que tenemos un conjunto diverso de usuarios entusiastas, y que aunque muchos siguen con la configuración clásica de Atlantis (un puñado de repos que ejecutan terraform contra AWS en GitHub), hay muchos casos de uso y direcciones diferentes hacia las que va la comunidad y en las que le gustaría ver que Atlantis diera soporte.

Estamos agradecidos con todos los que se tomaron el tiempo para compartir sus experiencias con Atlantis. Planeamos ejecutar este tipo de encuesta de manera semirregular, ¡manténganse atentos!

## Resultados anonimizados

### ¿Cómo interactúas con Atlantis?

![](../../../blog/2024/april-2024-survey-results/interact.webp)

Como era de esperar, la mayoría de los usuarios de Atlantis desempeñan múltiples funciones y están involucrados en todo el proceso de desarrollo.

### Cómo implementas Atlantis tú/tu organización

![](../../../blog/2024/april-2024-survey-results/deploy.webp)

La mayoría de los usuarios de terraform implementan usando Kubernetes y/o AWS. "Other Docker" usa docker pero no usa EKS ni Helm directamente, mientras que una minoría usa alguna otra combinación de tecnologías.

### ¿Qué herramienta(s) de Infrastructure as Code (IaC) usas con Atlantis?

![](../../../blog/2024/april-2024-survey-results/iac.webp)

La gran mayoría de los usuarios de Atlantis siguen usando terraform como alguna parte de su despliegue. Aproximadamente la mitad de ellos además usa Terragrunt, y OpenTofu parece estar ganando algo de terreno.

### ¿Cuántos repositorios administra tu Atlantis?

![](../../../blog/2024/april-2024-survey-results/repos.webp)

La mayoría de los usuarios tienen huellas relativamente modestas para administrar con Atlantis (aunque algunos monorepos grandes podrían quedar ocultos en los números).

### ¿Qué sistemas de control de versiones (VCSs) usas?

![](../../../blog/2024/april-2024-survey-results/vcs.webp)

La mayoría de los usuarios de Atlantis están usando GitHub, con una porción considerable en GitLab, seguida por Bitbucket y otros. Esto es análogo a las solicitudes de soporte y funcionalidades que los maintainers ven para los diversos VCSs en el código base.

### ¿Cuál es la funcionalidad más importante que consideras que falta en Atlantis?

![](../../../blog/2024/april-2024-survey-results/features.webp)

Como esta es una pregunta de formato libre, hubo una larga cola de respuestas, por lo que lo anterior solo muestra respuestas después de la normalización que tuvieron tres o más casos.

Drift Detection, así como las mejoras de infraestructura, fueron los ganadores obvios aquí. Después de eso, los usuarios se enfocaron en varias integraciones y mejoras de la UI.

## Conclusión

Siempre es interesante y emocionante para el equipo principal ver la amplitud del uso de Atlantis, y esperamos usar esta información para comprender las necesidades de la comunidad. ¡Atlantis siempre ha sido un esfuerzo impulsado por la comunidad, y esperamos poder seguir llevando ese espíritu hacia adelante!
