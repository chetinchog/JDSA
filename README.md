# JDSA - Job Data Scraping Assistant

JDSA es una herramienta de escritorio moderna diseñada para extraer y organizar información detallada de publicaciones de empleo en diversas plataformas. Ideal para analistas de reclutamiento y buscadores de empleo enfocados en la organización eficiente de datos.

## 🚀 Características Principales

- **Extracción Precisa**: Captura automáticamente datos clave como Título, Empresa, Ubicación, ID de Empleo y URL directa de postulación.
- **Búsqueda Masiva Inteligente**: Permite buscar y exportar grandes volúmenes de ofertas organizadas cronológicamente (las más recientes primero).
- **Control de Duplicados**: El sistema identifica y descarta automáticamente publicaciones repetidas o patrocinadas redundantes.
- **Detección de Vacantes Expiradas**: Identifica de forma proactiva empleos que ya no están disponibles, agilizando el proceso de selección.
- **Exportación en CSV**: Guarda los datos directamente en un archivo `.csv` estandarizado, listo para ser importado en bases de datos o planillas de cálculo.
- **Gestión Automatizada de Bloqueos**: Cuenta con mecanismos de prevención y reintentos automáticos para asegurar que tu búsqueda no se interrumpa.
- **Navegación Fluida**: Diseño de alto contraste con Modo Oscuro/Claro y previsualización completa de las descripciones del puesto sin salir de la app.

---

## 📖 Manual de Usuario Paso a Paso

### 1. Descarga e Instalación
- Dirígete a la sección de **Releases** en este repositorio.
- Descarga el archivo ejecutable (`jdsa.exe`) correspondiente a la última versión.
- Al ser una herramienta portable, no requiere instalación. Simplemente haz doble clic para **Abrir** la aplicación.

### 2. Inicio de Sesión / Registro
- Al abrir JDSA por primera vez, se te pedirá iniciar sesión.
- Si no posees una cuenta, presiona en **"¿No tenés cuenta? Solicitar acceso"**.
- Ingresa tu Email y una Contraseña, y haz clic en **Solicitar acceso**.
- Tu cuenta será creada, pero deberás esperar a que un **Administrador la habilite** manualmente para poder utilizar la herramienta.
- Una vez habilitado, simplemente ingresa tus credenciales en la pantalla principal para entrar. Puedes marcar la casilla **"Recordar mis datos localmente"** para no tener que escribirlos en tu próxima sesión.

### 3. Configuración y Carga de Cookies (Bypass de Seguridad)
Para evitar que las plataformas bloqueen tus búsquedas de forma prematura, es fundamental proveer una "Cookie de Sesión" fresca desde tu navegador:
1. Haz clic en el engranaje de **Configuración** (esquina superior derecha).
2. Abre tu navegador web principal y navega a la plataforma de búsqueda (ej. `ar.indeed.com`).
3. Presiona en JDSA el botón para **abrir el Login interactivo** y asegúrate de iniciar sesión en la plataforma y realizar al menos una búsqueda manual de prueba.
4. Con una extensión de navegador o desde las Herramientas de Desarrollador, copia el valor de la cookie y pégalo en el recuadro **"Session Cookie"** de JDSA.
5. Ajusta los **tiempos de espera (Espera entre páginas y empleos)**. Por defecto, recomendamos dejar pausas de entre **10 a 15 segundos** para evitar ser bloqueado o baneado por los sistemas anti-bot exhaustivos de plataformas como Indeed.

### 4. Realizar una Búsqueda
- En el menú lateral, selecciona la plataforma deseada.
- Ingresa tu término de búsqueda (ej. "Rust developer") y presiona **Buscar**.
- Durante la búsqueda, verás una barra de progreso. Si el flujo es interrumpido por medidas de seguridad de la plataforma, un cartel rojo te lo notificará y la aplicación esperará para que puedas actualizar tu cookie e intentar nuevamente gracias a su sistema de reintentos escalonados.

### 5. Ver Detalles del Empleo
- Una vez obtenidos los resultados, verás la lista de empleos en la pantalla principal.
- Haz clic en cualquier tarjeta de empleo para abrir un panel interactivo y leer cuidadosamente la **descripción completa**. Aquí podrás confirmar si los requerimientos de la oferta coinciden con tus expectativas antes de exportarla.

### 6. Exportar los Resultados
- Cuando la búsqueda termine (o si la cancelas voluntariamente para analizar lo ya encontrado), presiona el botón **Exportar TODO a CSV**.
- Elige una carpeta en tu computadora donde guardar el archivo.
- Los datos se exportarán listos, con campos pre-procesados (como el código de País ISO) en el estándar correcto para su uso posterior en tus procesos de recursos humanos.

---

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Consulta el archivo [LICENSE](LICENSE) para más detalles.
