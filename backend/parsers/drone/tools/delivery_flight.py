#!/usr/bin/env python3

import rospy
import logging
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from app.navigation.clover_navigation_controller import CloverNavigationController

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


def main():
    logger.info("="*60)
    logger.info("CLOVER DELIVERY MISSION")
    logger.info("="*60)
    
    map_file = Path(__file__).parent.parent / "config" / "aruco_map.txt"
    
    if not map_file.exists():
        logger.error(f"Map file not found: {map_file}")
        return False
    
    logger.info(f"Using map: {map_file}")
    
    nav = CloverNavigationController(str(map_file))
    
    logger.info("Initializing navigation system...")
    if not nav.initialize():
        logger.error("Navigation initialization failed")
        return False
    
    target_marker = 52
    logger.info(f"Target: Parcel Automat (Marker ID={target_marker})")
    
    try:
        success = nav.execute_delivery(target_marker)
        
        if success:
            logger.info("="*60)
            logger.info("DELIVERY COMPLETED SUCCESSFULLY")
            logger.info("="*60)
            
            rospy.sleep(5)
            logger.info("Mission complete. Ready for package drop.")
            
        else:
            logger.error("="*60)
            logger.error("DELIVERY FAILED")
            logger.error("="*60)
        
        return success
        
    except KeyboardInterrupt:
        logger.warning("Mission interrupted by user")
        return False
    except Exception as e:
        logger.error(f"Mission error: {e}")
        return False


if __name__ == "__main__":
    try:
        success = main()
        sys.exit(0 if success else 1)
    except Exception as e:
        logger.error(f"Fatal error: {e}")
        sys.exit(1)

