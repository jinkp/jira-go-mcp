# MVP — MCP Propio para Gestión de Releases en Jira

## 1. Objetivo

Crear un **MCP Server propio** especializado en la gestión de releases de Jira, complementando al MCP oficial de Atlassian.

El objetivo no es reemplazar el MCP oficial, sino cubrir funcionalidades avanzadas relacionadas con:

- Releases / Versions de Jira.
- Release Notes.
- Validaciones antes de deploy.
- Relación entre tickets, versiones y despliegues.
- Automatización de gobernanza técnica.
- Consultas inteligentes desde herramientas como Codex, OpenCode, Cursor, Claude Desktop u otros clientes MCP.

---

## 2. Problema que resuelve

Aunque Jira permite manejar releases mediante la entidad **Version**, muchas organizaciones necesitan reglas adicionales que no vienen listas en Jira o que no están cubiertas por el MCP oficial de Atlassian.

Ejemplos:

- Validar que un release tenga todos los tickets cerrados.
- Saber qué tickets quedaron pendientes.
- Generar release notes automáticamente.
- Marcar una versión como liberada desde un asistente.
- Consultar releases por producto, proyecto o fecha.
- Crear releases bajo una convención interna.
- Impedir despliegues si falta fixVersion.
- Preparar reportes para soporte, dirección o producto.

---

## 3. Alcance del MVP

### Incluido en el MVP

- Conexión con Jira Cloud REST API.
- Autenticación usando API Token o OAuth posteriormente.
- Crear releases en Jira.
- Actualizar releases existentes.
- Marcar releases como liberados.
- Archivar releases.
- Obtener releases por proyecto.
- Buscar release por nombre.
- Obtener issues relacionados a un release usando JQL.
- Generar release notes básicas.
- Validar si un release está listo para deploy.
- Exponer herramientas mediante MCP.

### Fuera del MVP inicial

- UI web completa.
- Integración profunda con Bitbucket.
- Integración con Azure DevOps Pipelines.
- OAuth multiusuario avanzado.
- Base de datos propia.
- Métricas históricas avanzadas.
- Aprobaciones formales tipo workflow.

---

## 4. Casos de uso principales

### Caso 1 — Crear release

El usuario escribe desde un cliente compatible con MCP:

```txt
Crea el release API v1.0.5 en el proyecto NSLICENSE con fecha 2026-05-30.
```

El MCP ejecuta:

- Valida que el proyecto exista.
- Valida que no exista una versión con el mismo nombre.
- Crea la versión en Jira.
- Devuelve confirmación con ID, nombre y fecha.

---

### Caso 2 — Consultar releases de un proyecto

```txt
Muéstrame los releases del proyecto NSLICENSE.
```

El MCP responde con:

- Nombre del release.
- Estado: liberado, no liberado o archivado.
- Fecha de inicio, si existe.
- Fecha de liberación.
- Número de issues relacionados.

---

### Caso 3 — Validar release antes de deploy

```txt
Valida si el release API v1.0.5 está listo para producción.
```

El MCP revisa:

- Issues asociados por fixVersion.
- Issues abiertos.
- Bugs críticos sin resolver.
- Tickets sin assignee.
- Tickets sin status Done.
- Issues bloqueados.

Resultado esperado:

```txt
El release no está listo para producción.

Problemas encontrados:
- 2 bugs siguen abiertos.
- 1 historia está en QA.
- 3 tickets no tienen evidencia registrada.
```

---

### Caso 4 — Generar release notes

```txt
Genera las release notes del release API v1.0.5.
```

El MCP consulta issues con:

```jql
project = NSLICENSE AND fixVersion = "API v1.0.5" ORDER BY issuetype, priority DESC
```

Y genera una salida en Markdown:

```md
# Release Notes — API v1.0.5

## Features
- NSL-120: Nueva validación de licencias.
- NSL-122: Mejora en endpoint de activación.

## Bugs corregidos
- NSL-130: Corrección en renovación de licencia.

## Hotfixes
- NSL-135: Ajuste urgente en expiración de token.
```

---

### Caso 5 — Marcar release como liberado

```txt
Marca el release API v1.0.5 como liberado con fecha de hoy.
```

El MCP:

- Busca la versión.
- Ejecuta validaciones opcionales.
- Actualiza `released = true`.
- Actualiza `releaseDate`.
- Devuelve confirmación.

---

## 5. Herramientas MCP propuestas

### 5.1 `jira_release_create`

Crea un release/version en Jira.

#### Input

```json
{
  "projectKey": "NSLICENSE",
  "name": "API v1.0.5",
  "description": "Release de mejoras para licenciamiento",
  "releaseDate": "2026-05-30",
  "startDate": "2026-05-20"
}
```

#### Output

```json
{
  "id": "10034",
  "name": "API v1.0.5",
  "projectKey": "NSLICENSE",
  "released": false,
  "archived": false,
  "releaseDate": "2026-05-30"
}
```

---

### 5.2 `jira_release_update`

Actualiza nombre, descripción, fechas o estado del release.

#### Input

```json
{
  "versionId": "10034",
  "name": "API v1.0.5",
  "description": "Release actualizado",
  "releaseDate": "2026-06-01"
}
```

---

### 5.3 `jira_release_mark_released`

Marca una versión como liberada.

#### Input

```json
{
  "projectKey": "NSLICENSE",
  "releaseName": "API v1.0.5",
  "releaseDate": "2026-05-30",
  "runValidation": true
}
```

---

### 5.4 `jira_release_archive`

Archiva una versión.

#### Input

```json
{
  "projectKey": "NSLICENSE",
  "releaseName": "API v1.0.5"
}
```

---

### 5.5 `jira_release_list_by_project`

Lista releases de un proyecto.

#### Input

```json
{
  "projectKey": "NSLICENSE",
  "includeArchived": false,
  "includeReleased": true
}
```

---

### 5.6 `jira_release_get_issues`

Obtiene issues asociados a un release.

#### Input

```json
{
  "projectKey": "NSLICENSE",
  "releaseName": "API v1.0.5",
  "fields": ["summary", "status", "assignee", "issuetype", "priority"]
}
```

---

### 5.7 `jira_release_generate_notes`

Genera release notes en Markdown.

#### Input

```json
{
  "projectKey": "NSLICENSE",
  "releaseName": "API v1.0.5",
  "groupBy": "issueType",
  "includeInternalComments": false,
  "format": "markdown"
}
```

---

### 5.8 `jira_release_validate_for_deploy`

Valida si un release está listo para desplegar.

#### Input

```json
{
  "projectKey": "NSLICENSE",
  "releaseName": "API v1.0.5",
  "rules": [
    "all_issues_done",
    "no_critical_bugs_open",
    "all_issues_have_assignee",
    "all_issues_have_fix_version"
  ]
}
```

#### Output

```json
{
  "ready": false,
  "releaseName": "API v1.0.5",
  "errors": [
    "Hay 2 issues que no están en Done.",
    "Hay 1 bug crítico abierto."
  ],
  "warnings": [
    "Hay 3 issues sin evidencia QA."
  ]
}
```

---

### 5.9 `jira_release_move_unresolved_issues`

Mueve issues no resueltos a otro release.

#### Input

```json
{
  "projectKey": "NSLICENSE",
  "fromRelease": "API v1.0.5",
  "toRelease": "API v1.0.6",
  "onlyUnresolved": true
}
```

---

## 6. Arquitectura propuesta

```txt
Cliente MCP
Codex / OpenCode / Cursor / Claude Desktop
        |
        v
MCP Server propio
        |
        |-- Tools MCP
        |-- Validaciones internas
        |-- Generador de release notes
        |-- Cliente Jira REST API
        |
        v
Jira Cloud REST API
```

---

## 7. Componentes técnicos

### 7.1 MCP Server

Responsable de exponer herramientas al cliente MCP.

Puede desarrollarse en:

- Node.js + TypeScript.
- Go.
- Python.

Recomendación para MVP:

```txt
Node.js + TypeScript
```

Motivos:

- Buen soporte para MCP SDK.
- Rápido para prototipar.
- Fácil manejo de JSON.
- Buen ecosistema HTTP.
- Compatible con despliegue en Docker.

---

### 7.2 Jira Client

Capa encargada de consumir Jira REST API.

Responsabilidades:

- Autenticación.
- Crear versions.
- Actualizar versions.
- Listar versions.
- Ejecutar JQL.
- Leer issues.
- Manejar errores de Jira.

---

### 7.3 Release Service

Capa de negocio del MCP.

Responsabilidades:

- Aplicar convenciones internas.
- Validar duplicados.
- Resolver projectKey a projectId.
- Buscar versionId por nombre.
- Preparar payloads para Jira.
- Agrupar issues.
- Generar release notes.

---

### 7.4 Validation Engine

Motor de reglas para validar releases.

Reglas iniciales:

```txt
all_issues_done
no_critical_bugs_open
all_issues_have_assignee
all_issues_have_fix_version
no_blocked_issues
no_unresolved_bugs
```

Reglas futuras:

```txt
qa_evidence_required
approved_pr_required
successful_build_required
security_review_required
product_owner_approval_required
```

---

### 7.5 Configuración

Variables de entorno:

```env
JIRA_BASE_URL=https://empresa.atlassian.net
JIRA_EMAIL=usuario@empresa.com
JIRA_API_TOKEN=token_seguro
JIRA_API_VERSION=3
DEFAULT_DONE_STATUSES=Done,Closed,Released
DEFAULT_CRITICAL_PRIORITIES=Highest,Critical
```

---

## 8. APIs de Jira necesarias

### Obtener proyecto

```http
GET /rest/api/3/project/{projectIdOrKey}
```

### Listar versiones de proyecto

```http
GET /rest/api/3/project/{projectIdOrKey}/versions
```

### Crear versión

```http
POST /rest/api/3/version
```

### Actualizar versión

```http
PUT /rest/api/3/version/{id}
```

### Eliminar versión

```http
DELETE /rest/api/3/version/{id}
```

### Buscar issues con JQL

```http
POST /rest/api/3/search/jql
```

### Obtener conteos relacionados a versión

```http
GET /rest/api/3/version/{id}/relatedIssueCounts
```

---

## 9. JQL base del MVP

### Issues por release

```jql
project = NSLICENSE AND fixVersion = "API v1.0.5" ORDER BY issuetype, priority DESC
```

### Issues no terminados

```jql
project = NSLICENSE AND fixVersion = "API v1.0.5" AND status NOT IN (Done, Closed, Released)
```

### Bugs abiertos

```jql
project = NSLICENSE AND fixVersion = "API v1.0.5" AND issuetype = Bug AND status NOT IN (Done, Closed, Released)
```

### Bugs críticos abiertos

```jql
project = NSLICENSE AND fixVersion = "API v1.0.5" AND issuetype = Bug AND priority IN (Highest, Critical) AND status NOT IN (Done, Closed, Released)
```

### Issues sin assignee

```jql
project = NSLICENSE AND fixVersion = "API v1.0.5" AND assignee IS EMPTY
```

### Issues bloqueados

```jql
project = NSLICENSE AND fixVersion = "API v1.0.5" AND issueLinkType = blocks
```

---

## 10. Convenciones sugeridas para releases

### Naming

```txt
{Producto} v{Major}.{Minor}.{Patch}
```

Ejemplos:

```txt
API v1.0.1
App v2.3.0
SRPayments v1.5.2
NSLicense API v1.0.4
```

### Tipos de release

```txt
Feature Release
Hotfix
Patch
Major Release
Internal Release
```

### Estados lógicos internos

Aunque Jira maneja `released` y `archived`, el MCP puede interpretar estados:

```txt
Planned
In Progress
Ready for QA
Ready for Deploy
Released
Archived
Blocked
```

Estos estados pueden calcularse a partir de issues, fechas y reglas internas.

---

## 11. Estructura propuesta del proyecto

```txt
jira-release-mcp/
├── src/
│   ├── index.ts
│   ├── server.ts
│   ├── config/
│   │   └── env.ts
│   ├── jira/
│   │   ├── jiraClient.ts
│   │   ├── jiraVersionsApi.ts
│   │   └── jiraSearchApi.ts
│   ├── releases/
│   │   ├── releaseService.ts
│   │   ├── releaseNotesService.ts
│   │   └── releaseValidationService.ts
│   ├── tools/
│   │   ├── jiraReleaseCreateTool.ts
│   │   ├── jiraReleaseUpdateTool.ts
│   │   ├── jiraReleaseListTool.ts
│   │   ├── jiraReleaseIssuesTool.ts
│   │   ├── jiraReleaseNotesTool.ts
│   │   └── jiraReleaseValidateTool.ts
│   ├── schemas/
│   │   └── releaseSchemas.ts
│   └── utils/
│       ├── logger.ts
│       └── errors.ts
├── tests/
│   ├── releaseService.test.ts
│   └── validationService.test.ts
├── .env.example
├── package.json
├── tsconfig.json
├── Dockerfile
└── README.md
```

---

## 12. Flujo interno de una tool MCP

Ejemplo: `jira_release_validate_for_deploy`

```txt
1. Recibir input del cliente MCP.
2. Validar schema del input.
3. Buscar proyecto en Jira.
4. Buscar versión por nombre.
5. Ejecutar JQL para obtener issues relacionados.
6. Ejecutar reglas de validación.
7. Construir resultado estructurado.
8. Responder al cliente MCP.
```

---

## 13. Seguridad

### Recomendaciones para MVP

- No guardar tokens en código.
- Usar `.env` local.
- Usar secretos en el ambiente de despliegue.
- Limitar permisos del usuario/token usado.
- Registrar logs sin exponer tokens.
- Validar inputs antes de consultar Jira.
- Evitar ejecutar JQL arbitrario sin control.

### Permisos mínimos recomendados

El token necesita permisos para:

- Leer proyectos.
- Leer issues.
- Leer versiones.
- Crear versiones.
- Editar versiones.
- Actualizar issues si se agregará asignación de `fixVersion`.

---

## 14. Ejemplo de interacción esperada

### Usuario

```txt
Valida el release NSLicense API v1.0.4 para producción.
```

### MCP

```json
{
  "tool": "jira_release_validate_for_deploy",
  "arguments": {
    "projectKey": "NSLICENSE",
    "releaseName": "NSLicense API v1.0.4",
    "rules": [
      "all_issues_done",
      "no_critical_bugs_open",
      "all_issues_have_assignee"
    ]
  }
}
```

### Respuesta

```txt
El release NSLicense API v1.0.4 no está listo para producción.

Errores:
- NSL-324 sigue en QA.
- NSL-330 es un bug crítico abierto.

Advertencias:
- NSL-318 no tiene assignee.

Recomendación:
Resolver los errores antes de marcar el release como liberado.
```

---

## 15. Roadmap sugerido

### Fase 1 — MVP funcional

- MCP server local.
- Jira API token.
- Crear release.
- Listar releases.
- Actualizar release.
- Obtener issues por release.
- Generar release notes básicas.
- Validar release para deploy.

### Fase 2 — Automatización avanzada

- Mover issues no resueltos a siguiente release.
- Asignar fixVersion a issues.
- Integrar Bitbucket PRs.
- Validar approvals.
- Validar build exitoso.
- Publicar release notes en Confluence.

### Fase 3 — Gobernanza DevOps

- Integrar Azure DevOps Pipelines.
- Relacionar commit SHA, build y release.
- Bloquear deploy si el release no cumple reglas.
- Auditoría de cambios.
- Notificaciones a Teams.

### Fase 4 — Portal visual

- Timeline de releases.
- Calendario por producto.
- Vista para soporte sin licencia Jira.
- Filtros por producto, equipo, tipo y fecha.
- Exportación a PDF/Markdown.

---

## 16. Criterios de éxito del MVP

El MVP se considera exitoso si permite:

- Crear un release desde un cliente MCP.
- Consultar releases de un proyecto.
- Ver issues asociados a un release.
- Generar release notes en Markdown.
- Validar si un release está listo para deploy.
- Marcar un release como liberado.
- Reducir consultas manuales dentro de Jira.
- Servir como base para automatizaciones futuras de DevOps.

---

## 17. Riesgos y consideraciones

### Riesgo 1 — Diferentes flujos por equipo

Cada equipo puede usar estados distintos como Done, Cerrado, Liberado o Finalizado.

Mitigación:

- Configurar estados válidos por proyecto.

---

### Riesgo 2 — FixVersion no siempre se usa bien

Si los equipos no asignan fixVersion, el MCP no tendrá buena visibilidad.

Mitigación:

- Crear validación de tickets sin fixVersion.
- Agregar reportes de higiene.

---

### Riesgo 3 — Permisos insuficientes

El token puede no tener permisos para crear o modificar versiones.

Mitigación:

- Crear una cuenta técnica con permisos controlados.
- Documentar permisos mínimos.

---

### Riesgo 4 — JQL diferente por proyecto

Algunos proyectos pueden tener campos personalizados.

Mitigación:

- Permitir configuración por proyecto.
- Crear presets de JQL.

---

## 18. Valor para la organización

Este MCP permitiría convertir Jira en una fuente operativa para automatización real de releases.

Beneficios:

- Menos trabajo manual.
- Mayor control antes de producción.
- Mejor visibilidad para soporte y dirección.
- Release notes automáticas.
- Reducción de errores en despliegues.
- Base para un portal interno de releases.
- Integración futura con Bitbucket, Azure DevOps y Teams.

---

## 19. Siguiente paso recomendado

Construir primero el MCP con estas 5 tools mínimas:

```txt
jira_release_create
jira_release_list_by_project
jira_release_get_issues
jira_release_generate_notes
jira_release_validate_for_deploy
```

Con esas herramientas ya se puede probar el valor del flujo completo:

```txt
Crear release → Asociar issues → Validar → Generar notas → Liberar
```

