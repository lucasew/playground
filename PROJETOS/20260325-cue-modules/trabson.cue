package workspaced

{
	modules: teste: #Module & {
		files: {
			".bashrc": {
				type: "lines"
				content: {
					"init2": "echo trabson"
				}
			}
		}
	}
}
