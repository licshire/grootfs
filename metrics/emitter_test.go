package metrics_test

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/grootfs/metrics"
	"code.cloudfoundry.org/grootfs/testhelpers"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/sonde-go/events"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Emitter", func() {
	Describe("TryEmitDuration", func() {
		var (
			fakeMetronPort   uint16
			fakeMetron       *testhelpers.FakeMetron
			fakeMetronClosed chan struct{}
			emitter          *metrics.Emitter
			logger           *lagertest.TestLogger
		)

		BeforeEach(func() {
			fakeMetronPort = uint16(5000 + GinkgoParallelNode())

			fakeMetron = testhelpers.NewFakeMetron(fakeMetronPort)
			Expect(fakeMetron.Listen()).To(Succeed())

			Expect(
				dropsonde.Initialize(fmt.Sprintf("127.0.0.1:%d", fakeMetronPort), "foo"),
			).To(Succeed())

			fakeMetronClosed = make(chan struct{})
			go func() {
				defer GinkgoRecover()
				Expect(fakeMetron.Run()).To(Succeed())
				close(fakeMetronClosed)
			}()

			emitter = metrics.NewEmitter()

			logger = lagertest.NewTestLogger("emitter")
		})

		AfterEach(func() {
			Expect(fakeMetron.Stop()).To(Succeed())
			Eventually(fakeMetronClosed).Should(BeClosed())
		})

		It("emits metrics", func() {
			emitter.TryEmitDuration(logger, "foo", time.Second)

			var fooMetrics []events.ValueMetric
			Eventually(func() []events.ValueMetric {
				fooMetrics = fakeMetron.ValueMetricsFor("foo")
				return fooMetrics
			}).Should(HaveLen(1))

			Expect(*fooMetrics[0].Name).To(Equal("foo"))
			Expect(*fooMetrics[0].Unit).To(Equal("nanos"))
			Expect(*fooMetrics[0].Value).To(Equal(float64(time.Second)))
		})
	})
})