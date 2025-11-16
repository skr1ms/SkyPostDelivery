import cv2
import numpy as np
from pathlib import Path
from reportlab.pdfgen import canvas
from reportlab.lib.pagesizes import A4
from reportlab.lib.units import cm
from reportlab.lib.utils import ImageReader
import io
from PIL import Image


A4_WIDTH, A4_HEIGHT = A4


def generate_marker(marker_id: int, marker_size: int = 200, dict_type=cv2.aruco.DICT_6X6_250):
    aruco_dict = cv2.aruco.getPredefinedDictionary(dict_type)
    marker_image = cv2.aruco.generateImageMarker(aruco_dict, marker_id, marker_size)
    return marker_image


def create_marker_with_border(marker_id: int, marker_size: int = 800, border_size: int = 100):
    marker = generate_marker(marker_id, marker_size)
    
    total_size = marker_size + 2 * border_size
    bordered_marker = np.ones((total_size, total_size), dtype=np.uint8) * 255
    
    bordered_marker[border_size:border_size + marker_size, 
                   border_size:border_size + marker_size] = marker
    
    text = f"ID: {marker_id}"
    font = cv2.FONT_HERSHEY_SIMPLEX
    font_scale = 2.5
    thickness = 5
    text_size = cv2.getTextSize(text, font, font_scale, thickness)[0]
    
    text_x = (total_size - text_size[0]) // 2
    text_y = total_size - border_size // 3
    
    cv2.putText(bordered_marker, text, (text_x, text_y), 
                font, font_scale, 0, thickness)
    
    return bordered_marker


def generate_markers_pdf(start_id: int, end_id: int, output_path: str, 
                         markers_per_page: int = 4):
    marker_ids = list(range(start_id, end_id + 1))
    num_pages = (len(marker_ids) + markers_per_page - 1) // markers_per_page
    
    pdf = canvas.Canvas(output_path, pagesize=A4)
    
    markers_per_row = 2
    marker_physical_size = 9 * cm
    
    x_margin = (A4_WIDTH - markers_per_row * marker_physical_size) / 3
    y_start = A4_HEIGHT - 3 * cm
    y_spacing = marker_physical_size + 2 * cm
    
    page_num = 1
    for page_start in range(0, len(marker_ids), markers_per_page):
        page_markers = marker_ids[page_start:page_start + markers_per_page]
        
        for idx, marker_id in enumerate(page_markers):
            row = idx // markers_per_row
            col = idx % markers_per_row
            
            marker_image = create_marker_with_border(marker_id, marker_size=800, border_size=100)
            
            img_buffer = io.BytesIO()
            pil_image = Image.fromarray(marker_image)
            pil_image.save(img_buffer, format='PNG')
            img_buffer.seek(0)
            
            x = x_margin + col * (marker_physical_size + x_margin)
            y = y_start - row * y_spacing
            
            img_reader = ImageReader(img_buffer)
            pdf.drawImage(img_reader, x, y, 
                         width=marker_physical_size, 
                         height=marker_physical_size)
        
        pdf.setFont("Helvetica", 10)
        pdf.drawString(A4_WIDTH - 5*cm, 1*cm, 
                      f"Page {page_num}/{num_pages} | IDs: {page_markers[0]}-{page_markers[-1]}")
        
        if page_start + markers_per_page < len(marker_ids):
            pdf.showPage()
        
        page_num += 1
    
    pdf.save()
    print(f"Saved PDF: {output_path}")
    print(f"  Markers ID {start_id}-{end_id}")
    print(f"  Pages: {num_pages}")
    print(f"  Marker size: 9x9 cm")


def generate_single_large_marker_pdf(marker_id: int, output_path: str):
    pdf = canvas.Canvas(output_path, pagesize=A4)
    
    marker_image = create_marker_with_border(marker_id, marker_size=1200, border_size=150)
    
    if marker_id == 52:
        text = "PARCEL AUTOMAT LANDING PAD"
        font = cv2.FONT_HERSHEY_SIMPLEX
        font_scale = 3.0
        thickness = 8
        text_size = cv2.getTextSize(text, font, font_scale, thickness)[0]
        text_x = (marker_image.shape[1] - text_size[0]) // 2
        text_y = 100
        cv2.putText(marker_image, text, (text_x, text_y), 
                   font, font_scale, 0, thickness)
    
    img_buffer = io.BytesIO()
    pil_image = Image.fromarray(marker_image)
    pil_image.save(img_buffer, format='PNG')
    img_buffer.seek(0)
    
    marker_size = 18 * cm
    x = (A4_WIDTH - marker_size) / 2
    y = (A4_HEIGHT - marker_size) / 2
    
    img_reader = ImageReader(img_buffer)
    pdf.drawImage(img_reader, x, y, width=marker_size, height=marker_size)
    
    pdf.setFont("Helvetica-Bold", 14)
    pdf.drawCentredString(A4_WIDTH / 2, y - 2*cm, f"ArUco Marker ID: {marker_id}")
    pdf.drawCentredString(A4_WIDTH / 2, y - 2.7*cm, "Physical size: 18x18 cm")
    
    pdf.save()
    print(f"Saved large marker PDF ID {marker_id}: {output_path}")
    print(f"  Physical print size: 18x18 cm")


def generate_simple_markers_pdf(marker_ids: list, output_path: str, marker_size_cm: float = 15.0):
    pdf = canvas.Canvas(output_path, pagesize=A4)
    
    marker_size = marker_size_cm * cm
    x = (A4_WIDTH - marker_size) / 2
    y = (A4_HEIGHT - marker_size) / 2 + 1.5*cm
    
    for idx, marker_id in enumerate(marker_ids):
        marker_image = create_marker_with_border(marker_id, marker_size=1000, border_size=120)
        
        img_buffer = io.BytesIO()
        pil_image = Image.fromarray(marker_image)
        pil_image.save(img_buffer, format='PNG', dpi=(300, 300))
        img_buffer.seek(0)
        
        img_reader = ImageReader(img_buffer)
        pdf.drawImage(img_reader, x, y, 
                     width=marker_size, height=marker_size,
                     preserveAspectRatio=True, anchor='c')
        
        pdf.setFont("Helvetica-Bold", 18)
        pdf.drawCentredString(A4_WIDTH / 2, y - 2.2*cm, f"ArUco Marker ID: {marker_id}")
        pdf.drawCentredString(A4_WIDTH / 2, y - 3*cm, f"Size: {marker_size_cm}x{marker_size_cm} cm")
        
        pdf.setFont("Helvetica", 10)
        pdf.drawString(x, y + marker_size + 0.3*cm, f"<-- {marker_size_cm} cm -->")
        
        pdf.setStrokeColorRGB(1, 0, 0)
        pdf.setLineWidth(1)
        pdf.line(x, y + marker_size + 0.1*cm, x + marker_size, y + marker_size + 0.1*cm)
        pdf.line(x, y + marker_size, x, y + marker_size + 0.2*cm)
        pdf.line(x + marker_size, y + marker_size, x + marker_size, y + marker_size + 0.2*cm)
        
        pdf.setStrokeColorRGB(0, 0, 0)
        
        ruler_x = 2*cm
        ruler_y = 2*cm
        pdf.setFont("Helvetica-Bold", 8)
        pdf.drawString(ruler_x, ruler_y + 0.7*cm, "RULER (check print scale):")
        pdf.setLineWidth(2)
        for i in range(11):
            pdf.line(ruler_x + i*cm, ruler_y, ruler_x + i*cm, ruler_y + 0.5*cm)
            if i % 5 == 0:
                pdf.setFont("Helvetica", 7)
                pdf.drawString(ruler_x + i*cm - 0.2*cm, ruler_y - 0.4*cm, f"{i}")
        pdf.line(ruler_x, ruler_y, ruler_x + 10*cm, ruler_y)
        pdf.setFont("Helvetica", 7)
        pdf.drawString(ruler_x + 10*cm + 0.2*cm, ruler_y - 0.1*cm, "10 cm")
        
        pdf.setFont("Helvetica", 10)
        pdf.drawString(ruler_x, 1.2*cm, f"Page {idx + 1}/{len(marker_ids)}")
        
        if idx < len(marker_ids) - 1:
            pdf.showPage()
    
    pdf.save()
    print(f"Saved PDF: {output_path}")
    print(f"  Markers: {marker_ids}")
    print(f"  Total pages: {len(marker_ids)}")
    print(f"  Each marker size: {marker_size_cm}x{marker_size_cm} cm")


def main():
    output_dir = Path("./aruco_markers")
    output_dir.mkdir(exist_ok=True, parents=True)
    
    print("=" * 60)
    print("ArUco Marker Generator - Simple Edition")
    print("=" * 60)
    print()
    
    marker_size_cm = 15.0
    marker_ids = list(range(0, 10)) + [52]
    
    output_path = str(output_dir / "aruco_markers_simple.pdf")
    generate_simple_markers_pdf(marker_ids, output_path, marker_size_cm)
    
    print()
    print("=" * 60)
    print("Done!")
    print("=" * 60)
    print()
    print("Generated markers:")
    print(f"  IDs: {marker_ids}")
    print(f"  File: aruco_markers_simple.pdf")
    print(f"  Pages: {len(marker_ids)} (1 marker per page)")
    print(f"  Marker size: {marker_size_cm}x{marker_size_cm} cm")
    print()
    print("Printing Instructions:")
    print("  1. Open PDF in Adobe Reader or similar")
    print("  2. Print at 100% scale (CRITICAL: NOT 'Fit to page'!)")
    print("  3. Check the RULER on each page - it should be exactly 10 cm")
    print("  4. If ruler is wrong, your printer settings are incorrect")
    print("  5. Laminate each page")
    print()
    print("Usage:")
    print("  - Markers 0-9: Place on floor every 2 meters")
    print("  - Marker 52: Place on TOP of parcel automat")
    print()
    print(f"Output directory: {output_dir.absolute()}")
    print()
    print("IMPORTANT: Don't forget to update config:")
    print(f"  MARKER_SIZE_CM={marker_size_cm}")
    print()


if __name__ == "__main__":
    main()

