using Plots, LinearAlgebra, Statistics, Printf 
plotly()
#--------------------------------------------------------------
#Dados a serem ajustados
Dados = [1 1509
2 1205
3 595
4 745
5 1605
6 1556
7 1548
8 1456
9 868
10 948]

X, Y = Dados[:,1], Dados[:,2]
n = length(X)
#--------------------------------------------------------------
#Implementando o modelo linear 
#Montando e resolvendo o sistema de equacoes normais 
A = [ones(n) X] 
beta = (A'*A)\A'*Y
#A solucao do sistema encontra-se na variavel beta 
phi1(x) = beta[1] + beta[2]*x 
#--------------------------------------------------------------
#Calculo do R2 
SSres = sum((Y - phi1.(X)).^2)
SStot = sum((Y .- mean(Y)).^2)
R2 = 1 - SSres/SStot

println("O coeficiente de determinacao do modelo phi1: ", R2)
#--------------------------------------------------------------
#plotando o grafico 
scatter(X,Y,label="Dados")
plot!(phi1,1:10,c=:orange,label="Ajuste")