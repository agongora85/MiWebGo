package constants

import "miweb/config"

const (
	URLBFF          = config.BFFLocal
	AuthCallback    = config.BFFQA + "/auth/callback"
	AuthCheck       = config.BFFQA + "/auth/check"
	AuthSession     = config.BFFQA + "/auth/session"
	AuthPermissions = config.BFFQA + "/auth/permissions"
	PermisosUsuario = config.BFFQA + "/api/PermisosUsuario"
	InicioSesion    = config.INICIOSESION
	TemplteIndex    = "template/index.html"
)
