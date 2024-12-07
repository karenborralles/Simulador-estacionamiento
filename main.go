package main

import (
	"simulador-final/models"
	"sync"

	"golang.org/x/image/font/basicfont"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
)

func run() {
	config := pixelgl.WindowConfig{
		Title:  "Simulador de Estacionamiento",
		Bounds: pixel.R(0, 0, 1024, 768), 
		VSync:  true,
	}
	ventana, err := pixelgl.NewWindow(config)
	if err != nil {
		panic(err)
	}

	atlasTexto := text.NewAtlas(basicfont.Face7x13, text.ASCII)

	estacionamiento := models.NuevoEstacionamiento(20) 
	var grupoEspera sync.WaitGroup

	estacionamiento.AgregarObservador(func() {
		ventana.Clear(colornames.Skyblue) 
		estacionamiento.Dibujar(ventana, atlasTexto)
		ventana.Update()
	})

	go estacionamiento.IniciarSimulacion(100, 0.5, &grupoEspera) 

	for !ventana.Closed() {
		ventana.Clear(colornames.Skyblue)
		estacionamiento.Dibujar(ventana, atlasTexto)
		ventana.Update()
	}

	grupoEspera.Wait()
}

func main() {
	pixelgl.Run(run)
}