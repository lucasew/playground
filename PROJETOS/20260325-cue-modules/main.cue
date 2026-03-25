package workspaced

{
	modules: teste: #Module & {
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
