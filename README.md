# Example of k8s operator build based on operator-sdk

https://blog.container-solutions.com/hs-fs/hubfs/kubernetes_operators_diagram1.png?width=1500&name=kubernetes_operators_diagram1.png

Primero inicializamos metadata
$  operator-sdk init --project-version=2 --domain example.org --license apache2 --owner "alknopfler" --repo github.com/alknopfler/example-operator-k8s
Generamos skaffolding
$ operator-sdk create api --group event-finder --version v1beta1 --kind SagaFinder

Lo primero que hacemos es montar el CRD como queremos tenerlo basado en type struct.
Para eso podemos modificar el yaml, pero en el operator-sdk definiendo el API/struct de tus datos, después pasas un make generate y make manifest y te generan ya los yaml de CRD

$ make generate && make manifest  -> con esto tenemos CRD yaml file en base a nuestro struct


Ahora implementamos el controller. para ello
Implmeentamos el reconciler loop con la lógica
Para ello usamos un spec de un template y además añadimos una vez que se ha creado el deployment un fichero de sincronización
