# todo

- [ ] properly support cancelling an in-flight deployment. need to somehow
      maintain some state or memory or something of the deployment and cancel the
      context if its cancelled. maybe a polling goroutine to see if it gets cancelled
      then cancel the context?
- [ ] pass the service definition or whatever directly to the kube deployment.
      this will help when adding new properties and all that
- [ ] make wingmen just another service definition. this is so that i dont need
      to specify a whole separate but the same config for wingmen, i can just reuse
      the same config and all that from services but for wingmen
