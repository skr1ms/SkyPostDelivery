import grpc


class OrchestratorGRPCClient:
    def __init__(self, orchestrator_url: str):
        self.orchestrator_url = orchestrator_url
        self.channel = None
        self.stub = None

    async def connect(self):
        try:
            options = [
                ('grpc.keepalive_time_ms', 30000),
                ('grpc.keepalive_timeout_ms', 5000),
                ('grpc.max_receive_message_length', 4 * 1024 * 1024),
            ]
            self.channel = grpc.aio.insecure_channel(
                self.orchestrator_url, options=options)

            from proto import orchestrator_pb2_grpc
            self.stub = orchestrator_pb2_grpc.OrchestratorServiceStub(
                self.channel)
            print("Connected to orchestrator gRPC")
        except Exception as e:
            print(f"Failed to connect to orchestrator: {e}")

    async def disconnect(self):
        if self.channel:
            await self.channel.close()


    async def request_cell_open(self, order_id: str, parcel_automat_id: str) -> dict:
        try:
            if not self.stub:
                await self.connect()

            from proto import orchestrator_pb2

            request = orchestrator_pb2.CellOpenRequest(
                order_id=order_id,
                parcel_automat_id=parcel_automat_id
            )

            response = await self.stub.RequestCellOpen(request, timeout=10.0)
            return {
                "success": response.success,
                "message": response.message,
                "cell_id": response.cell_id
            }
        except Exception as e:
            print(f"Failed to request cell open: {e}")
            return {
                "success": False,
                "message": str(e),
                "cell_id": ""
            }

