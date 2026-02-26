# JDSA - Job Description Scraping Assistant

JDSA es una herramienta de escritorio moderna diseñada para extraer y organizar información detallada de publicaciones de empleo de diversas plataformas.

## 🚀 Características

- **Extracción Precisa**: Captura Título, Empresa, Ubicación, ID de Empleo y URL de postulación.
- **Búsqueda Masiva**: Permite buscar resultados masivos e iterar sobre ellos, exportando el listado completo automáticamente. Ahora con ordenamiento cronológico (más recientes primero) y manejo seguro de búsquedas sin resultados.
- **Desduplicación Inteligente**: Descarta internamente registros orgánicos y patrocinados clonados para evitar basura en la exportación final.
- **Aislamiento de Errores**: Filtra vacantes que resultan inaccesibles durante operaciones masivas (404/Not Found) notificando un resumen limpio.
- **Descripción Completa**: Recupera la descripción íntegra del puesto manteniendo el formato y saltos de línea originales.
- **Modo Claro/Oscuro**: Tema oscuro con toggle, persiste la preferencia y detecta el tema del sistema.
- **Diseño Moderno**: Interfaz Material Design construida con Vue 3 y Tailwind CSS v4.
- **Navegación Fluida**: Diseño de alto fijo con scroll interno especializado para descripciones largas.
- **Exportación**: Guarda los datos scrapeados directamente en archivos JSON estructurados.
- **Seguridad**: Validación de URLs con protección contra SSRF (bloquea IPs privadas, localhost, esquemas no-HTTP).
- **Detección de Vacantes Expiradas**: Identifica automáticamente empleos que ya no están disponibles mediante señales de metadatos, selectores de alerta y datos estructurados, mostrando un badge de "Expirado".
- **Anti-Bloqueo**: Implementa headers dinámicos y respaldo en datos estructurados (JSON-LD) para máxima estabilidad.

## 🛠️ Tecnologías

- **Backend**: [Go 1.26](https://go.dev/) + [Colly](http://go-colly.org/)
- **Frontend**: [Vue.js 3](https://vuejs.org/) + [Tailwind CSS v4](https://tailwindcss.com/)
- **Desktop Framework**: [Wails v2](https://wails.io/)

## 📦 Instalación y Uso

### Ejecutable (Windows)
Este repositorio **no incluye el archivo .exe** por razones de tamaño y seguridad. Para obtenerlo, tenés dos opciones:
1. **Compilarlo vos mismo** siguiendo las instrucciones de abajo.
2. **Descargarlo** desde la sección de **Releases** de este repositorio (si están disponibles).

### Desarrollo y Compilación
Si deseás compilar el proyecto manualmente:

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
Herramienta eficiente para analistas de recruiting y buscadores de empleo enfocada en la organización de datos.
