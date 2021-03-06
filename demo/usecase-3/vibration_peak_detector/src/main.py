import os
import signal
import asyncio
import logging

import edgefarm_application as ef
from analyzer import Analyzer
from ads_event_reporter import AdsEventReporter

_logger = logging.getLogger(__name__)


async def main():
    loop = asyncio.get_event_loop()

    # Initialize EdgeFarm SDK
    if os.getenv("IOTEDGE_MODULEID") is not None:
        await ef.application_module_init_from_environment(loop)
    else:
        print("Warning: Running example outside IOTEDGE environment")
        await ef.application_module_init(
            loop, "SeatRes", "train-seat-info-monitor", "no-runtime-id"
        )

    # Create a queue that we will use to store events. Analyzer will report peaks there
    event_q = asyncio.Queue()

    # Connect to EdgeFarm service module mqtt-bridge
    mqtt_client = ef.AlmMqttModuleClient()

    # Prepare reporter to send detected peaks to ADS
    event_reporter = AdsEventReporter()

    # Start the Analyzer
    Analyzer(event_q, mqtt_client)

    def signal_handler():
        event_q.put_nowait("stop")

    for sig in ("SIGINT", "SIGTERM"):
        loop.add_signal_handler(getattr(signal, sig), signal_handler)

    while True:
        event = await event_q.get()
        _logger.info(f"main loop: event {event}")
        if type(event) is str and event == "stop":
            break
        else:
            await event_reporter.report(event)

    print("Shutting down...")
    await mqtt_client.close()
    await ef.application_module_term()


if __name__ == "__main__":
    logging.basicConfig(
        level=os.environ.get("LOGLEVEL", "INFO").upper(),
        format="%(asctime)s %(name)-12s %(levelname)-8s %(message)s",
        datefmt="%Y/%m/%d %H:%M:%S",
    )
    asyncio.run(main())
