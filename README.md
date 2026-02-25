# JDSA Scraper - Extractor de Empleos Argentina

JDSA (Job Description Scraper Argentina) es una herramienta de escritorio moderna diseñada para extraer información detallada de publicaciones de empleo en Indeed Argentina.

![JDSA Interface Header](file:///C:/Users/cheti/.gemini/antigravity/brain/8acd59d5-aeca-49f3-a7cc-8cf98190b58c/inspect_indeed_selectors_1771981445165.webp)

## 🚀 Características

- **Extracción Precisa**: Captura Título, Empresa, Ubicación, ID de Empleo y URL de postulación.
- **Descripción Completa**: Recupera la descripción íntegra del puesto manteniendo el formato original.
- **Diseño Moderno**: Interfaz Material Design construida con Vue 3 y Tailwind CSS v4.
- **Navegación Fluida**: Diseño de alto fijo con scroll interno especializado para descripciones largas.
- **Exportación**: Guarda los datos scrapeados directamente en archivos JSON.
- **Anti-Bloqueo**: Implementa headers dinámicos y respaldos en datos estructurados (JSON-LD) para evitar errores "Forbidden".

## 🛠️ Tecnologías

- **Backend**: [Go](https://go.dev/) + [Colly](http://go-colly.org/)
- **Frontend**: [Vue.js 3](https://vuejs.org/) + [Tailwind CSS v4](https://tailwindcss.com/)
- **Desktop Framework**: [Wails v2](https://wails.io/)

## 📦 Instalación y Uso

### Ejecutable (Windows)
1. Descarga el archivo `JDSA.exe` desde la carpeta `.\release`.
2. Ejecútalo directamente.

### Desarrollo y Compilación
Si deseas compilar el proyecto manualmente:

1. Asegúrate de tener instalado **Go**, **Node.js** y **Wails CLI**.
2. Clonar el repositorio.
3. Ejecutar el asistente de compilación:
   ```cmd
   build.bat
   ```
   *Esto generará el ejecutable en la carpeta `release`.*

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Consulta el archivo [LICENSE](LICENSE) para más detalles.

---
Desarrollado como una herramienta eficiente para analistas de recruiting y buscadores de empleo.
