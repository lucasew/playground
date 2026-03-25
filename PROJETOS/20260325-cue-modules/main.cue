package workspaced

{
	modules: teste: #Module & {
		args: message: "world"
		files: {
			".bashrc": {
				type: "lines"
				content: {
					"init": "echo hello"
				}
			}
		}
	}
}
