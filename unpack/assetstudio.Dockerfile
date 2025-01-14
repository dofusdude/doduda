FROM mcr.microsoft.com/dotnet/sdk:8.0

RUN apt-get update && apt-get install -y curl unzip wget git cmake build-essential

WORKDIR /app

RUN git clone --depth 1 https://github.com/aelurum/AssetStudio.git AssetStudioGit
RUN mkdir -p /app/AssetStudioGit/Texture2DDecoderNative/build
WORKDIR /app/AssetStudioGit/Texture2DDecoderNative/build
RUN cmake -DCMAKE_CXX_FLAGS="-fpermissive" ..
RUN make

# portable app from latest release
WORKDIR /app
RUN curl -s https://api.github.com/repos/aelurum/AssetStudio/releases/latest \
    | grep "browser_download_url.*AssetStudioModCLI_net8_portable.zip" \
    | cut -d : -f 2,3 \
    | tr -d \" \
    | wget -qi -

RUN unzip AssetStudioModCLI_net8_portable.zip -d AssetStudio

# copy native lib
RUN cp /app/AssetStudioGit/Texture2DDecoderNative/build/libTexture2DDecoderNative.so /app/AssetStudio/Texture2DDecoderNative.so

# cleanup
RUN rm /app/AssetStudioModCLI_net8_portable.zip
RUN rm -rf /app/AssetStudioGit

WORKDIR /app/AssetStudio

ENTRYPOINT ["dotnet", "AssetStudioModCLI.dll"]
