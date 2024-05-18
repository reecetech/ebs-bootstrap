```
$ sudo blockdev --getsize64 /dev/vdb
10737418240

$ sudo pvdisplay /dev/vdb --units B
...
  PV Size               10737418240 B  / not usable 4194304 B

$ sudo vgdisplay vgdemo --units B
...
  VG Size               10733223936 B

# 10737418240 - 10733223936 = 4194304 Bytes (4 KB)

$ sudo lvdisplay vgdemo/lvdemo --units B
...
LV Size                10733223936 B
```

```
LV -> VG
VG -> PV
PV -> Block Device
```

```
$ sudo lvs -o lv_name,vg_name,lv_size,vg_size,lv_attr --units B --nosuffix
  LV        VG        LSize        VSize        Attr      
  ubuntu-lv ubuntu-vg 31079792640B 31079792640B -wi-ao----
  lvdemo    vgdemo    10733223936B 10733223936B -wi-a-----
 
$ sudo vgs -o vg_name,pv_name --units B --nosuffix
  VG        PV        
  ubuntu-vg /dev/vda3 
  vgdemo    /dev/vdb

$ sudo pvs -o pv_name,pv_size,dev_size --units B --nosuffix
  PV         PSize        DevSize     
  /dev/vda3  31079792640B 31082938368B
  /dev/vdb   10733223936B 10737418240B
```