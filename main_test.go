package main

import "testing"

func TestDecodeUsuariosConComillasSimples(t *testing.T) {
	const entrada = `{'IdUsuario':'763','Nombre':'ALEJANDRO','Opciones':[{'IdOpcion':'1381','NombreObjeto':'Lectura de formulario','Permiso':'3'}]}`

	usuarios, err := decodeUsuarios([]byte(entrada), 0)
	if err != nil {
		t.Fatalf("decodeUsuarios() devolvió error: %v", err)
	}
	if len(usuarios) != 1 {
		t.Fatalf("se esperaba 1 usuario; se recibieron %d", len(usuarios))
	}
	if usuarios[0].IDUsuario != "763" || usuarios[0].Nombre != "ALEJANDRO" {
		t.Fatalf("datos del usuario incorrectos: %+v", usuarios[0])
	}
	if len(usuarios[0].Opciones) != 1 || usuarios[0].Opciones[0].NombreObjeto != "Lectura de formulario" {
		t.Fatalf("opciones incorrectas: %+v", usuarios[0].Opciones)
	}
}

func TestDecodeUsuariosCampoJsonComoTexto(t *testing.T) {
	const entrada = `{"json":"{'IdUsuario':'763','Nombre':'ALEJANDRO','Opciones':[]}"}`

	usuarios, err := decodeUsuarios([]byte(entrada), 0)
	if err != nil {
		t.Fatalf("decodeUsuarios() devolvió error: %v", err)
	}
	if len(usuarios) != 1 || usuarios[0].IDUsuario != "763" {
		t.Fatalf("resultado inesperado: %+v", usuarios)
	}
}
