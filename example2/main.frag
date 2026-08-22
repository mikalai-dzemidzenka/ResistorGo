
#version 330 core

in vec2 fragTexCoord;
out vec4 finalColor;

uniform sampler2D sceneTexture;
uniform sampler2D obstacleMask;
uniform vec2 screenSize;
uniform int lightCount;
uniform vec3 lights[8];
uniform vec3 lightColors[8];

void main()
{
    vec2 uv = vec2(fragTexCoord.x, 1.0 - fragTexCoord.y);
    vec3 scene = texture(sceneTexture, uv).rgb;
    vec3 ambient = vec3(0.08, 0.08, 0.10);
    vec3 lightAccum = vec3(0.0);

    for (int i = 0; i < 8; i++)
    {
        if (i >= lightCount) break;

        vec2 lightPos = lights[i].xy / screenSize;
        float radius = lights[i].z;
        vec2 delta = uv - lightPos;
        float dist = length(delta) * screenSize.x;

        float attenuation = clamp(1.0 - dist / radius, 0.0, 1.0);
        if (attenuation <= 0.0) continue;

        float shadow = 1.0;
        const int steps = 32;
        for (int s = 1; s <= steps; s++)
        {
            float t = float(s) / float(steps);
            vec2 sampleUV = mix(lightPos, uv, t);
            if (texture(obstacleMask, sampleUV).r > 0.5)
            {
                shadow = 0.0;
                break;
            }
        }

        lightAccum += lightColors[i] * attenuation * shadow;
    }

    vec3 totalLight = ambient + lightAccum;
    totalLight = min(totalLight, vec3(1.0));
    finalColor = vec4(scene * totalLight, 1.0);
}