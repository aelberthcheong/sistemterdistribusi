## Latar Belakang

Pengolahan data dalam jumlah besar membutuhkan waktu lama jika dikerjakan oleh satu proses saja. Salah satu solusinya adalah membagi pekerjaan ke beberapa proses yang berjalan bersamaan pada sistem terdistribusi.

Proyek ini mengimplementasikan **MapReduce** dengan bahasa **Go**. Satu *coordinator* membagikan tugas *map* dan *reduce* ke beberapa *worker* melalui RPC. Coordinator juga memberikan ulang tugas ke worker lain jika tugas belum selesai dalam 10 detik, sehingga hasil tetap benar walaupun ada worker yang gagal. Aplikasi yang dijalankan adalah penghitung kata (`wc`) dan indeks kata (`indexer`) pada 8 berkas teks buku dari Project Gutenberg.

Rancangan mengacu pada [Lab 1: MapReduce MIT 6.5840](https://pdos.csail.mit.edu/6.824/labs/lab-mr.html) dan makalah MapReduce. Proyek ini dibuat untuk tugas simulasi sistem terdistribusi di Universitas Mikroskil.

Cara Run & Build
```go
// build the plugins first

go run . coordinator /var/tmp/mr-sock pg-*.txt
go run . worker ./wc.so /var/tmp/mr-sock
go run . sequential ./wc.so pg-*.txt
```
