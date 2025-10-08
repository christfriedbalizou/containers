#=========== STAGE: BASE =====================================================#

ARG PYTHON_VERSION=3.14-bookworm
FROM python:$PYTHON_VERSION AS base

ENV PYTHONUNBUFFERED=1
ENV DOCKERIZE_VERSION=v0.6.1

# Define the Git tag to clone. Renovate will update this value.
ARG VERSION

RUN set -eux;

RUN mkdir -p /etc/apt/keyrings;

# Install common dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    sudo \
    ca-certificates \
    curl \
    gnupg \
    gettext \
    locales \
    rsync \
    && rm -rf /var/lib/apt/lists/*

# Define Locale
RUN localedef -i en_US -c -f UTF-8 -A /usr/share/locale/locale.alias en_US.UTF-8
ENV LANG=en_US.utf8

# Add dockerize (for healthchecks in docker-compose, if still desired)
RUN curl -L https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz | tar xzf - -C /usr/local/bin

COPY . /

#=========== STAGE: DEVELOPMENT_BASE =========================================#

FROM base AS development_base

# Add build dependencies, including libpq-dev for psycopg2 compilation
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    libpq-dev \
    git \
    npm \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Set working directory to /code where the Mathesar source will be cloned
WORKDIR /code/

# Clone the Mathesar repository and checkout the specific tag
# This makes `requirements-dev.txt` available for the pip install step
# --depth 1 for shallow clone to reduce image size
# --branch $GIT_TAG ensures it checks out the specified tag
# . clones into the current directory (/code/)
RUN git clone --depth 1 --branch $VERSION https://github.com/mathesar-foundation/mathesar.git . \
    && rm -rf .git # Remove the .git directory to reduce the final image size

# --- NEW: Remove any Dockerfiles from the cloned source code ---
# This ensures that only your custom Dockerfile is defining the image build.
# It uses `find` to locate files named "Dockerfile" or starting with "Dockerfile."
# and deletes them within the cloned repository.
RUN find . -type f -name "Dockerfile*" -delete

# Install python dependencies for local development
# `requirements-dev.txt` is now available because the repository has been cloned
RUN pip install --no-cache-dir -r requirements-dev.txt

# Compile translation files
RUN python manage.py compilemessages

# Add NodeJS signing key and source
RUN curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg && \
    chmod 644 /etc/apt/keyrings/nodesource.gpg && \
    echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_$NODE_MAJOR.x nodistro main" \
    | tee /etc/apt/sources.list.d/nodesource.list

# Install node
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    nodejs \
    && rm -rf /var/lib/apt/lists/*


#=========== STAGE: DEVELOPMENT ==============================================#

FROM development_base AS development

# Install npm packages
# Path adjusted because current WORKDIR is /code/
RUN cd mathesar_ui && npm ci && cd ..

EXPOSE 8000 3000 6006

CMD ["bash", "./bin/mathesar_dev"]


#=========== STAGE: PRE_PRODUCTION ===========================================#

FROM development_base AS pre_production

# Source code is already cloned in DEVELOPMENT_BASE, so no need to clone again.
WORKDIR /code/ # Ensure working directory is /code for the packaging script

# Run the packaging script on the already cloned source code
RUN python3 ./build-scripts/package/package.py


#=========== STAGE: PRODUCTION ===============================================#

FROM base AS production

# Copy packaged source files from the pre_production stage
COPY --from=pre_production /code/dist/mathesar.tar.gz .
RUN tar -xzf mathesar.tar.gz && rm mathesar.tar.gz
WORKDIR /code/mathesar

# --- NEW STAGE FOR S3 AND OIDC INTEGRATION (custom_production) ---
# This stage uses the build stage's output for the final slim image
FROM base AS custom_production

# Mathesar's application code is typically located at /code/mathesar in their images.
WORKDIR /code/mathesar

# Install additional Python dependencies for S3 storage and OIDC
# These are added to the environment provided by Mathesar's base image.
# COPY --from=base /custom-requirements.txt .
# RUN pip install --no-cache-dir -r ./custom-requirements.txt
# RUN rm -f ./custom-requirements.txt

# Copy your custom settings file into the image
# Ensure the path aligns with where Mathesar expects settings, or where you define it.
COPY --from=base /custom_settings.py /code/mathesar/config/settings/custom_settings.py

# Set the default settings module for the runtime.
ENV DJANGO_SETTINGS_MODULE=config.settings.custom_settings

# --- FINAL STAGE: Production S3/OIDC Image ---
FROM base AS final
# Copy the modified application code (with S3 and OIDC packages and custom settings)
# from the custome_production stage into the final image.
COPY --from=custom_production /code/mathesar /code/mathesar

WORKDIR /code/mathesar
# The original entrypoint and CMD from the base_mathesar image will be inherited.
# You typically don't need to redefine CMD here unless you want to override Mathesar's default.
