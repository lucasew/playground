package workspaced

{
	modules: teste: #Module & {
		args: {
			message: string
		}
		files: {
			".bashrc": {
				type: "lines"
				content: {
					"init2": "echo \(args.message)"
				}
			}
		}
	}
}
