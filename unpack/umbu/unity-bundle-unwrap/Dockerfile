FROM mcr.microsoft.com/dotnet/sdk:9.0

WORKDIR /app
COPY *.csproj ./
RUN dotnet restore
COPY . ./
RUN dotnet publish -c Release -o out

CMD ["dotnet", "out/unity-bundle-unwrap.dll"]