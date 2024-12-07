package models

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

type Vehiculo struct {
	ID           int           
	EspacioUsado *Espacio    
	TiempoParqueo time.Duration 
}

func CrearVehiculo(id int) *Vehiculo {
	return &Vehiculo{
		ID:           id,
		EspacioUsado: nil,
		TiempoParqueo: time.Second * time.Duration(30+rand.Intn(15)), 
	}
}

func GenerarVehiculos(estacionamiento *Estacionamiento, totalVehiculos int, lambda float64, grupoEspera *sync.WaitGroup) {
	rand.Seed(time.Now().UnixNano()) 
	for i := 1; i <= totalVehiculos; i++ {
		vehiculo := CrearVehiculo(i) 
		grupoEspera.Add(1)
		go vehiculo.Mover(estacionamiento, grupoEspera) 

		intervalo := generarIntervaloPoisson(lambda)
		time.Sleep(intervalo) 
	}
	grupoEspera.Wait()
}

func generarIntervaloPoisson(lambda float64) time.Duration {
	u := rand.Float64()
	intervalo := -math.Log(1-u) / lambda
	return time.Duration(intervalo * float64(time.Second))
}

func (v *Vehiculo) Mover(estacionamiento *Estacionamiento, grupoEspera *sync.WaitGroup) {
	defer grupoEspera.Done()

	for {
		estacionamiento.AgregarMensaje(fmt.Sprintf("Vehiculo %d intentando estacionarse...", v.ID))

		estacionamiento.CanalPuerta <- struct{}{} 
		espacio := estacionamiento.OcuparEspacio(v.ID)
		<-estacionamiento.CanalPuerta 

		if espacio != nil {
			v.EspacioUsado = espacio
			estacionamiento.AgregarMensaje(fmt.Sprintf("Vehiculo %d estacionado en el espacio %d", v.ID, espacio.Numero))

			time.Sleep(v.TiempoParqueo) 

			estacionamiento.LiberarEspacio(espacio) 
			estacionamiento.AgregarMensaje(fmt.Sprintf("Vehiculo %d dejó el espacio %d", v.ID, espacio.Numero))
			return
		}

		estacionamiento.AgregarMensaje(fmt.Sprintf("Vehiculo %d esperando por un espacio...", v.ID))
		time.Sleep(time.Second)
	}
}
