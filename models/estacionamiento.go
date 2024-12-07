package models

import (
	"fmt"
	"sync"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
)

type Espacio struct {
	Numero       int  
	EstaOcupado  bool 
	IDVehiculo   int 
}

type Estacionamiento struct {
	Espacios            []*Espacio      
	Observadores        []func()        
	TotalVehiculos      int            
	VehiculosParqueados int           
	Mensajes            []string       
	CanalPuerta         chan struct{}   
	CanalEspacios       chan struct{}   
	mutex               sync.Mutex     
}

func NuevoEstacionamiento(capacidad int) *Estacionamiento {
	espacios := make([]*Espacio, capacidad)  //se inicializa
	for i := 0; i < capacidad; i++ {
		espacios[i] = &Espacio{Numero: i + 1, EstaOcupado: false, IDVehiculo: -1} //tiene un numero único y no hay vehiculo
	}																		

	return &Estacionamiento{
		Espacios:      espacios,
		CanalPuerta:   make(chan struct{}, 1),         
		CanalEspacios: make(chan struct{}, capacidad),
	}
}
 
func (e *Estacionamiento) AgregarObservador(observador func()) { //función observadora a la lista de observadores
	e.Observadores = append(e.Observadores, observador)
}

func (e *Estacionamiento) NotificarObservadores() { //funciones observadoras
	for _, observador := range e.Observadores {
		observador()
	}
}

func (e *Estacionamiento) IniciarSimulacion(totalVehiculos int, lambda float64, grupoEspera *sync.WaitGroup) {
	e.TotalVehiculos = totalVehiculos  //asigna el número total de vehículos a simular (100)

	for range e.Espacios {      //Llena
		e.CanalEspacios <- struct{}{}
	}

	GenerarVehiculos(e, totalVehiculos, lambda, grupoEspera)
}

func (e *Estacionamiento) OcuparEspacio(idVehiculo int) *Espacio {
	<-e.CanalEspacios 

	e.mutex.Lock() //Bloquea hasta que haya un espacio disponible (CanalEspacios)
	defer e.mutex.Unlock()

	for _, espacio := range e.Espacios { //Busca un espacio libre, lo ocupa y notifica a los observadores
		if !espacio.EstaOcupado {
			espacio.EstaOcupado = true
			espacio.IDVehiculo = idVehiculo
			e.VehiculosParqueados++
			e.NotificarObservadores()
			return espacio
		}
	}
	return nil
}

func (e *Estacionamiento) LiberarEspacio(espacio *Espacio) {
	e.mutex.Lock()
	defer e.mutex.Unlock()	//Protege la operación de liberar un espacio con un mutex.

	espacio.EstaOcupado = false
	espacio.IDVehiculo = -1	//Marca el espacio como libre, devuelve el token al canal CanalEspacios y notifica a los observadores
	e.VehiculosParqueados--
	e.CanalEspacios <- struct{}{}
	e.NotificarObservadores()
}

func (e *Estacionamiento) AgregarMensaje(mensaje string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.Mensajes = append(e.Mensajes, mensaje)
	if len(e.Mensajes) > 10 {	//Limita el historial a los últimos 10 mensajes y notifica los cambios.
		e.Mensajes = e.Mensajes[1:]
	}
	e.NotificarObservadores()
}

func (e *Estacionamiento) Dibujar(ventana pixel.Target, atlas *text.Atlas) {
	imd := imdraw.New(nil)

	for _, espacio := range e.Espacios {
		posX := float64((espacio.Numero-1)%5)*120.0 + 250.0
		posY := 500.0 - float64((espacio.Numero-1)/5)*100.0

		if espacio.EstaOcupado {
			imd.Color = pixel.RGB(0, 0, 1)
		} else {
			imd.Color = pixel.RGB(0.8, 0.8, 0.8)
		}

		imd.Push(pixel.V(posX, posY))
		imd.Push(pixel.V(posX+80, posY+60))
		imd.Rectangle(0)
	}

	imd.Draw(ventana)

	texto := text.New(pixel.V(50, 700), atlas)
	fmt.Fprintf(texto, "Total Vehiculos: %d\n", e.TotalVehiculos)
	fmt.Fprintf(texto, "Vehiculos Estacionados: %d\n", e.VehiculosParqueados)
	fmt.Fprintf(texto, "Lugares Disponibles: %d\n", len(e.Espacios)-e.VehiculosParqueados)
	texto.Draw(ventana, pixel.IM.Scaled(texto.Orig, 2))

	terminal := text.New(pixel.V(50, 200), atlas)
	for i, mensaje := range e.Mensajes {
		terminal.Color = colornames.White
		terminal.Dot = pixel.V(50, 200-float64(i*20))
		fmt.Fprintln(terminal, mensaje)
	}
	terminal.Draw(ventana, pixel.IM.Scaled(terminal.Orig, 1.5))
}
