; File: millionare.ahk
; Contents: dual-color ball generator
; Author: chanfung
; Revision:
;   03/27/2009 create
;   03/28/2009 add function: send to clipboard

#NoTrayIcon
#SingleInstance force

; create gui interface
Gui, Add, Edit, section w30 vNum0 center
Gui, Add, Edit, w30 xp+35 vNum1 center
Gui, Add, Edit, w30 xp+35 vNum2 center
Gui, Add, Edit, w30 xp+35 vNum3 center
Gui, Add, Edit, w30 xp+35 vNum4 center
Gui, Add, Edit, w30 xp+35 vNum5 center
Gui, Add, Edit, w30 xp+35 vNum6 center cRed
Gui, Add, Button, xs90 gGenerate vOKButton, Generate!
Gui, +AlwaysOnTop 
Gui, Show
GuiControl, Focus, OKButton 
return

; generate numbers, each is different from 
; each other
Generate:
index = 0
loop, 6
{
    ; get a random number different from numbers 
    ; generated before
    loop
    {
        Random, temp, 1, 33
        cmp_index = 0
        loop
        {
            if ( cmp_index = index or Num%cmp_index% = temp )
                break
            cmp_index += 1
        }
        if ( cmp_index = index )
            break       
    }
    Num%index% := temp
    index += 1
}

; sort the six numbers
cnt = 5
loop, 6
{
    index1 = 0
    index2 = 1  
    loop, %cnt%
    {
        value1 := Num%index1%
        value2 := Num%index2%
        if ( value1 > value2 )
        {
            Num%index1% := value2
            Num%index2% := value1
        }
        index1 += 1
        index2 += 1
    }
    cnt -= 1
}

; generate the special number
Random, Num6, 1, 16

; display the numbers
index = 0
clipboard =
loop, 7
{
    value := Num%index%
    GuiControl, Text, Num%index%, %value%
    index += 1
    clipboard = %clipboard% %value%
}
; send to clipboard
clipboard = %clipboard%`r`n
return

GuiClose:
ExitApp
return
