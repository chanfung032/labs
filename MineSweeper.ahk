;
; AutoHotkey Version: 1.x
; Language:       English
; Platform:       Win9x/NT/xp
; Author:         Fung Chan <chanfung032@gmail.com>
;
; Script Function:
;   Winmine Automaton
; Revision:
;   3/4/2009 created
;

; #persistent

; check if the winmine.exe exist or not
; if it does, activate its main window
Process, Exist, winmine.exe
if (ErrorLevel = 0)
{
    MsgBox, Winmine doesnot exist!
    Exit
}
pid := ErrorLevel
WinGet, wid, ID, ahk_pid %pid%
WinActivate, ahk_id %wid%
WinWaitActive, ahk_id %wid%

; get the level of the game
WinGetPos, , , width, height, ahk_id %wid%
xDimension := Round((width - 26) / 16)
yDimension := Round((height - 110) / 16)

; change the coordinate mode from 
; screen to active window
CoordMode, Mouse, Relative
CoordMode, Pixel, Relative

; init
PixelSearch, px, , 0, 120, 20, 120, 0x808080
PixelSearch, , py, 15, 93, 15, 101, 0x808080
left := px + 3
top := py + 3
; MsgBox, %xDimension% %yDimension% %left% %top%

; start playing by randomly click a ceil
; Random, xIndex, 0, xDimension - 1
; Random, yIndex, 0, yDimension - 1
; ClickCeil(xIndex, yIndex, "L")

loop 
{
done := 0
loop %yDimension%
{
    rowIndex := A_Index - 1
    
    loop %xDimension%
    {
        columnIndex := A_Index - 1
        ceilStatus := GetCeilStatus(rowIndex, columnIndex)
        
        ; if the ceil is not a number, skip it
        if (ceilStatus < 1) or (ceilStatus > 8)
            continue
        
        flagNum = 0
        unsweepedNum = 0
        xOffset = -1
        loop 3
        {
            yOffset = -1
            loop 3
            {
                ; omit the center ceil
                if (xOffset = 0) and (yOffset = 0)
                {
                    yOffset += 1
                    Continue
                }
                ; caculate the number of the mine which is unsweeped
                ; or is flaged
                yCmpIndex := rowIndex + yOffset
                xCmpIndex := columnIndex + xOffset
                if (xCmpIndex >= 0) and (xCmpIndex < xDimension)
                    if (yCmpIndex >= 0) and (yCmpIndex < yDimension)
                    {
                        tempStatus := GetCeilStatus(yCmpIndex, xCmpIndex) 
                        if (tempStatus = 15)
                        {
                            xArray%unsweepedNum% := xCmpIndex
                            yArray%unsweepedNum% := yCmpIndex
                            unsweepedNum += 1
                        }
                        else if (tempStatus = 14)
                        {
                            flagNum += 1
                        }
                    }
                yOffset += 1
            }
            xOffset += 1
        }
        
        if (unsweepedNum = 0)
            continue
        if (ceilStatus = unsweepedNum + flagNum)
        {
            index = 0
            loop %unsweepedNum%
            {
                ClickCeil(yArray%index%, xArray%index%, "R")
                index += 1
            }
            done += 1
        }
        else if (ceilStatus = flagNum)
        {
            index = 0
            loop %unsweepedNum%
            {
                ClickCeil(yArray%index%, xArray%index%, "L")
                index += 1
            }
            done += 1
        }
    }
}

; MsgBox, %done%
if (done = 0)
{
    break
}
    
}

Exit

; click the ceil
ClickCeil(rowIndex, columnIndex, button)
{
    global left
    global top
    global wid
    
    xCoord := left + columnIndex * 16
    yCoord := top + rowIndex * 16
    ControlClick, x%xCoord% y%yCoord%, ahk_id %wid%, , %button%, , NA
}
    

; get the current status of the ordered ceil
; 0 : blank / sweeped
; 1 ~ 8: number
; 15 : blank / sweeping
GetCeilStatus(rowIndex, columnIndex)
{
    global left
    global top
    
    xCoord := left + 16 * columnIndex
    yCoord := top + 16 * rowIndex
    
    PixelGetColor, ceilColor, xCoord, yCoord
    if ( ceilColor = 0xFFFFFF )
    {
        ; if the ceil is not sweeped
        PixelGetColor, ceilColor, xCoord + 8, yCoord + 7
        if ( ceilColor = 0xC0C0C0 )
            status = 15
        else 
            status = 14
    }
    else
    {
        PixelGetColor, ceilColor, xCoord + 7, yCoord + 6
        if ( ceilColor = 0xFFFFFF )
        {
            ; if is a bomb
            status = 10
        }
        else
        {
            ; if the ceil is a number or blank 
            PixelGetColor, ceilColor, xCoord + 9, yCoord + 8
            if ( ceilColor = 0xFF0000 )
                status = 1
            else if ( ceilColor = 0x008000 )
                status = 2
            else if ( ceilColor = 0x0000FF )
                status = 3
            else if ( ceilColor = 0x800000 )
                status = 4
            else if ( ceilColor = 0x000080 )
                status = 5
            else if ( ceilColor = 0x808000 )
                status = 6
            else if ( ceilColor = 0x000000 )
                status = 7
            else if ( ceilColor = 0x808080 )
                status = 8
            else
                status = 0
        }
    }
        
    return status
}








