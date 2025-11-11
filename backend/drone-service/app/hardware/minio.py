import io
from datetime import datetime
from minio import Minio
from minio.error import S3Error
from config.config import settings


class MinIOClient:
    def __init__(self):
        self.client = Minio(
            settings.MINIO_ENDPOINT,
            access_key=settings.MINIO_ROOT_USER,
            secret_key=settings.MINIO_ROOT_PASSWORD,
            secure=settings.MINIO_USE_SSL
        )
        self.bucket_name = settings.MINIO_BUCKET_RECORDS
    
    async def upload_frame(self, drone_id: str, delivery_id: str, frame_data: bytes, frame_number: int) -> str | None:
        try:
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            object_name = f"{drone_id}/{delivery_id}/{timestamp}_frame_{frame_number}.jpg"
            
            stream = io.BytesIO(frame_data)
            self.client.put_object(
                self.bucket_name,
                object_name,
                stream,
                length=len(frame_data),
                content_type="image/jpeg"
            )
            return object_name
        except S3Error as e:
            print(f"Error uploading frame: {e}")
            return None
