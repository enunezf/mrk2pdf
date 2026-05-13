# Citas, enlaces e imágenes

## Citas (blockquotes)

> Cualquier tonto puede escribir código que una computadora pueda entender.
> Los buenos programadores escriben código que los humanos puedan entender.
>
> — Martin Fowler

## Citas anidadas

> Primer nivel de cita.
>
> > Segundo nivel anidado dentro del primero.
> >
> > > Tercer nivel, aún más adentro.
>
> De vuelta al primer nivel.

## Cita con formato interno

> **Importante**: este bloque combina *énfasis*, `código en línea` y un
> [enlace al repositorio](https://github.com/enunez/mrk2pdf) dentro de la cita.

## Enlaces

Visita la [documentación oficial de Go](https://go.dev/doc/) o consulta el
[spec de CommonMark](https://spec.commonmark.org/).

Enlaces con título: [Anthropic](https://www.anthropic.com "Página principal").

URLs automáticas (autolinks): <https://example.com> y <mailto:correo@example.com>.

## Enlaces por referencia

Este es un enlace por [referencia][1], y este es [otro][repo].
También funciona con [el mismo texto][el mismo texto] como etiqueta.

[1]: https://example.com
[repo]: https://github.com/enunez/mrk2pdf "Repositorio mrk2pdf"
[el mismo texto]: https://example.org

## Imágenes

Imagen inline:

![Logo de Go](https://go.dev/images/go-logo-white.svg "Logo de Go")

Imagen por referencia:

![Gopher][gopher-ref]

[gopher-ref]: https://go.dev/images/gophers/ladder.svg "Gopher escalando"

## Notas al pie (si la librería las soporta)

Aquí hay una frase con una nota al pie[^nota1] y otra más adelante[^nota2].

[^nota1]: Esta es la primera nota al pie.
[^nota2]: Esta nota es más extensa, con varias líneas y un enlace a
    [Wikipedia](https://es.wikipedia.org).
