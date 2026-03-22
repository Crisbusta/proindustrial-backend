package repository

import "github.com/crisbusta/proindustrial-backend-public/internal/model"

var sharedMaquinariasChildren = []model.SubSubcategory{
	{Slug: "arriendo", Name: "Arriendo"},
	{Slug: "venta", Name: "Venta"},
}

var sharedAsesoriasChildren = []model.SubSubcategory{
	{Slug: "inspeccion", Name: "Inspección"},
	{Slug: "calificacion", Name: "Calificación"},
	{Slug: "informatica", Name: "Informática"},
	{Slug: "contable", Name: "Contable"},
}

var CategoryGroups = []model.CategoryGroup{
	{
		Slug:        "termofusion",
		Name:        "Termofusión",
		Description: "Unión de tuberías PEAD y polipropileno mediante calor. Para proyectos de agua potable, riego y minería.",
		Icon:        "pipe",
		Subcategories: []model.Subcategory{
			{Slug: "distribuidoras", Name: "Distribuidoras Tubos", Description: "Proveedores y distribuidores de tuberías PEAD, polipropileno y accesorios.", Icon: "building"},
			{Slug: "maquinarias", Name: "Maquinarias", Description: "Arriendo y venta de máquinas de termofusión y electrofusión.", Icon: "wrench", Children: sharedMaquinariasChildren},
			{Slug: "repuestos", Name: "Repuestos", Description: "Repuestos y accesorios para máquinas de termofusión.", Icon: "package"},
			{Slug: "servicios", Name: "Servicios", Description: "Servicios de movimiento de tierra, ingeniería y flete especializados.", Icon: "network", Children: []model.SubSubcategory{
				{Slug: "movimiento-tierra", Name: "Movimiento de Tierra"},
				{Slug: "ingenieria", Name: "Ingeniería"},
				{Slug: "flete", Name: "Flete"},
			}},
			{Slug: "asesorias", Name: "Asesorías", Description: "Asesorías técnicas, inspección, calificación e informática.", Icon: "shield", Children: sharedAsesoriasChildren},
		},
	},
	{
		Slug:        "geomembranas",
		Name:        "Geomembranas",
		Description: "Instalación de liners y geomembranas HDPE para impermeabilización de tranques, piscinas y rellenos.",
		Icon:        "layers",
		Subcategories: []model.Subcategory{
			{Slug: "distribuidoras", Name: "Distribuidoras Membrana", Description: "Proveedores y distribuidores de geomembranas HDPE y accesorios.", Icon: "building"},
			{Slug: "maquinarias", Name: "Maquinarias", Description: "Arriendo y venta de extrusoras y equipos de instalación de geomembranas.", Icon: "wrench", Children: sharedMaquinariasChildren},
			{Slug: "repuestos", Name: "Repuestos", Description: "Repuestos y equipos de medición para instalación de geomembranas.", Icon: "package"},
			{Slug: "servicios", Name: "Servicios", Description: "Servicios de ingeniería, movimiento de tierra y flete especializados.", Icon: "network", Children: []model.SubSubcategory{
				{Slug: "ingenieria", Name: "Ingeniería"},
				{Slug: "movimiento-tierra", Name: "Movimiento Tierra"},
				{Slug: "flete", Name: "Flete"},
			}},
			{Slug: "asesorias", Name: "Asesorías", Description: "Asesorías técnicas, inspección, calificación e informática.", Icon: "shield", Children: sharedAsesoriasChildren},
		},
	},
}

var Regions = []string{
	"Todas las regiones",
	"Arica y Parinacota",
	"Tarapacá",
	"Antofagasta",
	"Atacama",
	"Coquimbo",
	"Valparaíso",
	"Metropolitana",
	"O'Higgins",
	"Maule",
	"Ñuble",
	"Biobío",
	"La Araucanía",
	"Los Ríos",
	"Los Lagos",
	"Aysén",
	"Magallanes",
}
