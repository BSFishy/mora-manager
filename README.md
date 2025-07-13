# mora manager

> [!NOTE]
> this project is pretty much sunset. i have learned what i needed to learn from
> it and now moving onto a different architecture. just using helm and deploying
> in phases.
>
> the main learning was k8s manifests and playing around with language design
> and implementation. i got what i wanted and continuing to maintain this would
> be a lot of effort for the same benefit of what i'll get with the new system.

this is the home of the mora manager. this is a service that lives on the k8s
cluster and manages actually actioning the configuration. it will take in a
configuration as input then expose a web endpoint to allow ad hoc configuration
for things like grabbing api keys and whatnot.
