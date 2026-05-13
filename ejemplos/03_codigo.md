# Bloques de código

## Código en línea

Usar `fmt.Println` para imprimir. La variable `cfg.Output` se construye a partir de `cfg.Input`.
También se puede invocar con `./mrk2pdf -i archivo.md -t default`.

## Bloque sin lenguaje

```
$ ./mrk2pdf -i ejemplo.md
Archivo de salida configurado en: ejemplo.pdf
Creando carpeta de plantillas...
Usando template: default
```

## Bloque con Go

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "uso: cmd <arg>")
        os.Exit(1)
    }
    fmt.Println("hola,", os.Args[1])
}
```

## Bloque con Bash

```bash
#!/usr/bin/env bash
set -euo pipefail

for archivo in *.md; do
    echo "Procesando: $archivo"
    ./mrk2pdf -i "$archivo"
done
```

## Bloque con JSON

```json
{
  "name": "mrk2pdf",
  "version": "0.1.0",
  "templates": ["default", "informe", "carta"]
}
```

## Bloque con YAML

```yaml
template: default
output:
  format: pdf
  margins:
    top: 2.5cm
    bottom: 2.5cm
```

## Líneas largas (overflow)

```
Esta es una línea muy larga que debería probar el comportamiento de overflow horizontal en el bloque de código preformateado del PDF final cuando supera el ancho de la página A4.
```
