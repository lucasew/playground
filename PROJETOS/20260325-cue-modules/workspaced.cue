package workspaced

#ArtifactBase: {
	type: string
	...
}

#ArtifactDownload: #ArtifactBase & {
	type: "download"
	url:  string
	hash: string
}

#ArtifactText: #ArtifactBase & {
	type:    "text"
	content: string
}

#ArtifactLines: #ArtifactBase & {
	type: "lines"
	content: [string]: string
}

#Artifact: #ArtifactDownload | #ArtifactText | #ArtifactLines

#Module: {
	files: [string]: #Artifact
}

{
	modules: [string]: #Module
}
