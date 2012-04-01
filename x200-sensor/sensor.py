"""
This script is a python wrapper for the ThinkPad Accelerometer Interface
which is part of the thinkpad's active protection system. It has only one
method named fetch and will return a tuple indicates the current status
of the accelerometer.  
"""

from ctypes import *

class SensorData(Structure):
    _fields_ = [('present_state', c_int),
                ('raw_accel_x', c_short),
                ('raw_accel_y', c_short),
                ('accel_x', c_short),
                ('accel_y', c_short),
                ('temperature', c_char),
                ('zero_g_x', c_short),
                ('zero_g_y', c_short)]
    
sensor_dll = windll.LoadLibrary("C:/Windows/System32/Sensor.DLL")
sensor_data = SensorData()

def fetch():
    sensor_dll.ShockproofGetAccelerometerData(byref(sensor_data))
    rv = (sensor_data.present_state,
          sensor_data.raw_accel_x,
          sensor_data.raw_accel_y,
          sensor_data.accel_x,
          sensor_data.accel_y,
          sensor_data.temperature,
          sensor_data.zero_g_x,
          sensor_data.zero_g_y)
    return rv
