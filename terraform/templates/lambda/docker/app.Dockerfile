FROM public.ecr.aws/lambda/provided:al2
WORKDIR /app
COPY bootstrap main
