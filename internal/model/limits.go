package model

// Límites canónicos de los formularios públicos.
//
// Fuente única de verdad del backend. Su equivalente en el cliente es
// frontend-public/src/constants/limits.ts, y los dos deben moverse juntos.
// El tope de descripción (600) coincide con el que el panel del proveedor
// ya aplicaba al mismo campo, para que registro y edición no se contradigan.
const (
	MinCompanyName = 3
	MaxCompanyName = 120

	MaxEmail = 150

	MinPhone = 7
	MaxPhone = 30

	MaxRegion = 100

	MaxServices    = 10
	MaxServiceSlug = 40

	MinDescription = 40
	MaxDescription = 600

	// Motivo de rechazo que escribe el admin en el backoffice.
	MaxRejectionReason = 500

	// MaxSlugBase deja margen dentro de companies.slug VARCHAR(100)
	// para el sufijo "-NN" que desambigua nombres repetidos.
	MaxSlugBase = 80

	// MaxBodyBytes acota cualquier cuerpo JSON entrante. El formulario
	// más grande de la aplicación no llega a 4 KB.
	MaxBodyBytes = 1 << 20 // 1 MB
)
