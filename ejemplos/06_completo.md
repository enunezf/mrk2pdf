# Informe técnico: migración de servicio de autenticación

**Autor:** Equipo de Plataforma
**Fecha:** 2026-05-13
**Estado:** Borrador para revisión

---

## Resumen ejecutivo

Este documento describe la migración del servicio de autenticación desde la
implementación legada en *Node.js* hacia el nuevo servicio en **Go**, con foco en
**reducción de latencia**, **simplificación operativa** y **cumplimiento de los
requisitos de auditoría** establecidos por el equipo legal.

> **Decisión clave**: la migración se hará por fases, manteniendo ambos servicios
> en paralelo durante 30 días para permitir rollback rápido si surgen incidentes.

## Objetivos

1. Reducir la latencia p99 de autenticación de **180 ms a menos de 50 ms**.
2. Eliminar la dependencia de `redis-cluster-legacy`, ya fuera de soporte.
3. Centralizar el manejo de tokens en un único servicio.
4. Habilitar trazabilidad completa para auditoría.

### Métricas objetivo

| Métrica            | Actual    | Objetivo  | Mejora |
|--------------------|----------:|----------:|-------:|
| Latencia p50       |    65 ms  |    20 ms  |   69 % |
| Latencia p99       |   180 ms  |    50 ms  |   72 % |
| Tasa de error      |   0.45 %  |   0.05 %  |   89 % |
| Costo mensual (USD)|    1 200  |      400  |   67 % |

## Arquitectura propuesta

### Componentes principales

- **API Gateway**
  - Termina TLS
  - Reenvía a `auth-svc` vía gRPC
- **`auth-svc` (Go)**
  - Valida credenciales contra base de datos primaria
  - Emite y verifica JWT
  - Publica eventos de auditoría
- **Base de datos**
  - PostgreSQL 15 con réplica de lectura
  - Migración automática vía `goose`
- **Cache de sesiones**
  - Redis 7, single-instance con persistencia AOF

### Flujo simplificado

```
[Cliente] --HTTPS--> [API Gateway] --gRPC--> [auth-svc] --SQL--> [PostgreSQL]
                                                  |
                                                  +--SET--> [Redis]
                                                  |
                                                  +--EVENT--> [Audit Bus]
```

## Plan de migración

### Fase 1 — Preparación (semanas 1-2)

- [x] Validar diseño con el equipo de seguridad
- [x] Provisionar infraestructura en *staging*
- [ ] Implementar suite de pruebas de carga
- [ ] Documentar plan de rollback

### Fase 2 — Despliegue en sombra (semanas 3-4)

El nuevo servicio recibirá tráfico real en modo *shadow*: las respuestas se
comparan con el servicio legado pero no se devuelven al cliente. Esto permite
detectar discrepancias **sin riesgo para el usuario**.

```bash
# Habilitar modo shadow
kubectl set env deploy/auth-svc SHADOW_MODE=true
kubectl rollout status deploy/auth-svc
```

### Fase 3 — Corte progresivo (semanas 5-6)

Rampa de tráfico controlada usando feature flags:

| Día | Porcentaje hacia `auth-svc` | Criterio de avance         |
|----:|---------------------------:|----------------------------|
|   1 |                       5 %  | Error rate < 0.1 %         |
|   3 |                      25 %  | Latencia p99 < 60 ms       |
|   7 |                      50 %  | Sin alertas críticas 48 h  |
|  14 |                     100 %  | Aprobación del *on-call*   |

### Fase 4 — Decomisionamiento (semana 7+)

Apagar el servicio legado solo tras **30 días** sin tráfico real.

## Riesgos identificados

> **Riesgo alto**: la migración de tokens existentes requiere que ambos servicios
> compartan la clave de firma durante la transición. Cualquier filtración
> compromete a ambos sistemas simultáneamente.

1. **Pérdida de sesiones activas** durante el corte
   - *Mitigación*: los JWT emitidos por el servicio legado seguirán siendo
     válidos hasta su expiración natural.
2. **Diferencias semánticas** en la validación de contraseñas
   - *Mitigación*: la suite de *shadow testing* compara byte-a-byte las
     respuestas durante 2 semanas antes de cualquier corte.
3. **Latencia inesperada** bajo carga real
   - *Mitigación*: prueba de carga con tráfico sintético al 200 % del pico
     histórico antes de la Fase 3.

## Decisiones técnicas

### ¿Por qué Go?

- Compilación a binario único, despliegue simple
- Concurrencia nativa con *goroutines*
- Ecosistema maduro de bibliotecas criptográficas (`crypto/*`, `golang.org/x/crypto`)
- Equipo de plataforma ya tiene experiencia en `inventory-svc` y `billing-svc`

### ¿Por qué no Rust?

Considerado seriamente, pero descartado por:

1. Curva de aprendizaje del equipo (~6 meses estimados)
2. Tiempos de compilación incompatibles con el ciclo de CI actual
3. Beneficios marginales sobre Go para este caso de uso

## Referencias

- [JWT Best Practices (RFC 8725)](https://datatracker.ietf.org/doc/html/rfc8725)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- Documento interno: `Política de gestión de secretos v3.2`
- Ticket de seguimiento: `PLAT-1842`

---

*Este documento sigue activo hasta el cierre de la Fase 4. Cualquier cambio
requiere aprobación del equipo de plataforma y revisión de seguridad.*
