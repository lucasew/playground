namespace PSModule

open System.Management.Automation

[<Cmdlet("Get", "Hello")>]
type GetHelloCmdlet () =
    inherit PSCmdlet ()

    [<Parameter(HelpMessage = "Name to greet")>]
    member val Name : string = "" with get, set

    override this.EndProcessing () =
        let text = match this.Name with
                   | "" -> "Hello, world!"
                   | name -> "Hello, "+ name + "!"
        this.WriteObject ( text )
        base.EndProcessing ()
