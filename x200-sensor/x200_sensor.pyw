import sensor
import Tkinter

WIN_WIDTH = 640
WIN_HEIGHT = 400
REFRESH_RATE = 100

def refresh_canvas():
    global vx, vy, x, y
    sensor_status = sensor.fetch()
    vx = sensor_status[4] - 496
    vy = sensor_status[3] - 528
    x = min(max(x + int(vx * 0.5), 10), WIN_WIDTH-10)
    y = min(max(y + int(vy * 0.5), 10), WIN_HEIGHT-10)
    
    items = canvas.find_all()
    canvas.delete(items)
    canvas.create_oval(x-10, y-10, x+10, y+10, fill='yellow')
    canvas.after(REFRESH_RATE, refresh_canvas)
   
root = Tkinter.Tk()
root.title('X200 Sensor')
canvas = Tkinter.Canvas(root, width=WIN_WIDTH, height=WIN_HEIGHT, bg='white')
canvas.pack()
canvas.after(REFRESH_RATE, refresh_canvas)

vx, vy = 0, 0
x, y = WIN_WIDTH/2, WIN_HEIGHT/2

root.mainloop()
