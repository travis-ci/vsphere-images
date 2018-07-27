package vsphereimages

import (
	"context"
	"net/url"
	"sync"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25/progress"
)

// copied from the real progress logger for vsphere-images commands
// which in turn is copied from the govc progress logger
//
// this one is modified to not print anything
// the Sinker interface is surprisingly complicated to implement correctly

type progressLogger struct {
	wg sync.WaitGroup

	sink chan chan progress.Report
	done chan struct{}
}

func newProgressLogger() *progressLogger {
	p := &progressLogger{
		sink: make(chan chan progress.Report),
		done: make(chan struct{}),
	}

	p.wg.Add(1)

	go p.loopA()

	return p
}

func (p *progressLogger) loopA() {
	var err error

	defer p.wg.Done()

	for stop := false; !stop; {
		select {
		case ch := <-p.sink:
			err = p.loopB(ch)
			if err != nil {
				stop = true
			}
		case <-p.done:
			stop = true
		}
	}
}

func (p *progressLogger) loopB(ch <-chan progress.Report) error {
	var r progress.Report
	var ok bool
	var err error

	for ok = true; ok; {
		select {
		case r, ok = <-ch:
			if !ok {
				break
			}
			err = r.Error()
		}
	}

	return err
}

func (p *progressLogger) Sink() chan<- progress.Report {
	ch := make(chan progress.Report)
	p.sink <- ch
	return ch
}

func (p *progressLogger) Wait() {
	close(p.done)
	p.wg.Wait()
}

type SimulatedService struct {
	Model  *simulator.Model
	Server *simulator.Server
}

func StartService() (*SimulatedService, error) {
	model := simulator.VPX()
	model.Machine = 4
	model.ClusterHost = 6
	if err := model.Create(); err != nil {
		return nil, err
	}
	s := model.Service.NewServer()

	service := &SimulatedService{
		Model:  model,
		Server: s,
	}
	return service, nil
}

func (service *SimulatedService) Stop() {
	service.Server.Close()
	service.Model.Remove()
}

func (service *SimulatedService) URL() *url.URL {
	return service.Server.URL
}

func (service *SimulatedService) NewClient(ctx context.Context) (*govmomi.Client, error) {
	return govmomi.NewClient(ctx, service.URL(), false)
}

func (service *SimulatedService) NewFinder(ctx context.Context) (*find.Finder, error) {
	client, err := service.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return find.NewFinder(client.Client, false), nil
}
